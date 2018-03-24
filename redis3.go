package redis3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Client defines a RediS3 client
type Client struct {
	svc     *s3.S3
	opt     *Options
	timeout int

	ctx context.Context
}

// KeyMetadata defines RediS3 key matadata which provides some useful
// information about the stored object.
type KeyMetadata struct {
	ValueType  string
	Locked     bool
	ExpireTime int64
	LastUpdate int64
}

// NewClient returns a RediS3 client.
func NewClient(opt *Options) (*Client, error) {
	var c Client
	var err error

	err = opt.init()
	if err != nil {
		return nil, err
	}

	awsConfig := aws.Config{Region: aws.String(opt.Region)}

	if opt.Endpoint != "" {
		awsConfig.S3ForcePathStyle = aws.Bool(true)
		awsConfig.WithEndpoint(opt.Endpoint)
	}

	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}

	c.svc = s3.New(sess)
	c.opt = opt
	c.ctx = context.Background()

	err = c.checkS3Client()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
