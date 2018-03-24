package redis3

import "errors"

// Options defines RediS3 client option
type Options struct {
	// AWS S3 bucket to be selected as database persistent storage.
	Bucket string

	// When AutoCreateBucket is set to true the client will try to create the
	// the AWS S3 bucket if it doesn't exist.
	// Of course, AWS credentials/instance profile requires proper permissions
	// so the client can create the bucket in S3.
	// Default is false.
	AutoCreateBucket bool

	// AWS Region where S3 bucket is located/created.
	Region string

	// AWS API Endpoint
	Endpoint string

	// Timeout defines S3 upload timeout in seconds.
	// If not defined or set to 0, there is no timeout.
	Timeout int

	// ReadOnly enables read only queries.
	// Default is false.
	readOnly bool

	// EnforceConsistency enables key locking while modifying keys.
	// This increases the number of requests to AWS S3 API and decreases
	// performance signlificantly but consistency is warranted when several
	// clients are modifying same key.
	// Default is false.
	EnforceConsistency bool
}

func (opt *Options) init() error {
	if opt.Bucket == "" {
		return errors.New("client: no bucket specified in options")
	}

	if opt.Region == "" {
		return errors.New("client: no region specified in options")
	}

	return nil
}
