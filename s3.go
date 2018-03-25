package redis3

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type tag struct {
	key   string
	value string
}

// Avoid padding characters in base64 output
var rawStdEncoding = b64.StdEncoding.WithPadding(b64.NoPadding)

func metadataToTags(metadata *KeyMetadata) string {
	var tags string

	tags = fmt.Sprintf("ValueType=%s&ExpireTime=%d&LastUpdate=%d&Locked=false",
		rawStdEncoding.EncodeToString([]byte(metadata.ValueType)),
		metadata.ExpireTime,
		metadata.LastUpdate,
	)

	return tags
}

func gobDecoder(body []byte, object interface{}) error {
	var err error

	dec := gob.NewDecoder(bytes.NewReader(body))
	err = dec.Decode(object)

	return err
}

func gobEndcoder(value interface{}) ([]byte, error) {
	var err error
	var buf bytes.Buffer

	gob.Register(value)
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Client) checkS3Client() error {
	output, err := c.svc.ListBuckets(nil)
	if err != nil {
		return err
	}
	for _, bucket := range output.Buckets {
		if *bucket.Name == c.opt.Bucket {
			return nil
		}
	}
	// If AutoCreateBucket is set to true, try to create AWS S3 Bucket.
	if !c.opt.AutoCreateBucket {
		return errors.New("client: specified bucket does not exist")
	} else {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(c.opt.Timeout)*time.Second,
		)
		defer cancel()

		_, err = c.svc.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(c.opt.Bucket),
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: aws.String(c.opt.Region),
			},
		},
		)
	}
	return err
}

func (c *Client) deleteKey(key string) error {
	var err error

	if c.opt.EnforceConsistency {
		err = c.lockKey(key)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.opt.Timeout)*time.Second,
	)
	defer cancel()

	_, err = c.svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.opt.Bucket),
		Key:    aws.String(key),
	})

	return err
}

func (c *Client) downloadKey(key string, value interface{}) (KeyMetadata, error) {
	var metadata KeyMetadata
	var err error

	metadata, err = c.getKeyMetadata(key)
	if err != nil {
		return metadata, err
	}
	// key does not exist or metadata tagging has been deleted and can't be
	// processed properly
	if metadata.LastUpdate == 0 {
		err = errors.New("key not found")
		return metadata, err
	}

	// If key is expired, key is deleted and "key not found" error returned
	if metadata.ExpireTime < time.Now().Unix() && metadata.ExpireTime != 0 {
		err = c.deleteKey(key)
		// If EnforceConsistency is set and key is already locked might be being
		// modified by another process/client (come back later)
		if err != nil {
			return metadata, err
		}
		err = errors.New("key not found")
		return metadata, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.opt.Timeout)*time.Second,
	)
	defer cancel()

	var response *s3.GetObjectOutput
	response, err = c.svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.opt.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return metadata, err
	}

	var b []byte
	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return metadata, err
	}

	err = gobDecoder(b, value)

	return metadata, err
}

func (c *Client) uploadKey(key string, value interface{}, metadata KeyMetadata) error {
	var err error

	if c.opt.EnforceConsistency {
		err = c.lockKey(key)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.opt.Timeout)*time.Second,
	)
	defer cancel()

	var b []byte
	b, err = gobEndcoder(value)
	if err != nil {
		return err
	}

	body := bytes.NewReader(b)
	tags := metadataToTags(&metadata)

	_, err = c.svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(c.opt.Bucket),
		Key:     aws.String(key),
		Body:    body,
		Tagging: &tags,
	})

	return err
}

func (c *Client) getKeyMetadata(key string) (KeyMetadata, error) {
	var metadata KeyMetadata
	var err error

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.opt.Timeout)*time.Second,
	)
	defer cancel()

	keyTagging, err := c.svc.GetObjectTaggingWithContext(
		ctx,
		&s3.GetObjectTaggingInput{
			Bucket: aws.String(c.opt.Bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return metadata, nil
			default:
				return metadata, err
			}
		}
	}

	for _, tag := range keyTagging.TagSet {
		switch *tag.Key {
		case "ValueType":
			var decodedValueType []byte
			decodedValueType, err = rawStdEncoding.DecodeString(*tag.Value)
			if err != nil {
				return metadata, err
			}
			metadata.ValueType = string(decodedValueType)
		case "Locked":
			metadata.Locked, err = strconv.ParseBool(*tag.Value)
		case "ExpireTime":
			metadata.ExpireTime, err = strconv.ParseInt(*tag.Value, 10, 64)
		case "LastUpdate":
			metadata.LastUpdate, err = strconv.ParseInt(*tag.Value, 10, 64)
		}
		if err != nil {
			return metadata, err
		}
	}

	return metadata, err
}

func (c *Client) setTags(key string, tags []tag) error {
	var err error
	var s3TagSet []*s3.Tag

	for _, tag := range tags {
		s3Tag := s3.Tag{Key: &tag.key, Value: &tag.value}
		s3TagSet = append(s3TagSet, &s3Tag)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(c.opt.Timeout)*time.Second,
	)
	defer cancel()

	_, err = c.svc.PutObjectTaggingWithContext(
		ctx,
		&s3.PutObjectTaggingInput{
			Bucket: aws.String(c.opt.Bucket),
			Key:    aws.String(key),
			Tagging: &s3.Tagging{
				TagSet: s3TagSet,
			},
		},
	)

	return err
}

func (c *Client) unlockKey(key string) error {
	return c.lockKeySwitch(key, false)
}

func (c *Client) lockKey(key string) error {
	return c.lockKeySwitch(key, true)
}

func (c *Client) lockKeySwitch(key string, lock bool) error {
	metadata, err := c.getKeyMetadata(key)
	if err != nil {
		return err
	}

	if lock {
		switch metadata.Locked {
		case true:
			// If key has expired we can proceed key even if is already locked
			if metadata.ExpireTime >= time.Now().Unix() || metadata.ExpireTime == 0 {
				err = errors.New("key already locked")
			}
		case false:
			// If LastUpdate == 0 means that key doesn't exist so, nothing to lock
			if metadata.LastUpdate > 0 {
				lockedTag := tag{key: "Locked", value: "true"}
				err = c.setTags(key, []tag{lockedTag})
			}
		}
	} else {
		switch metadata.Locked {
		case true:
			lockedTag := tag{key: "Locked", value: "false"}
			err = c.setTags(key, []tag{lockedTag})
		}
	}

	return err
}
