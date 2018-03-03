package redis3

import "errors"

// Options defines RediS3 client option
type Options struct {
	// AWS S3 bucket to be selected as database persistent storage.
	Bucket string

	// AWS Region where S3 bucket is created.
	Region string

	// AWS API Endpoint
	Endpoint string

	// Timeout defines S3 upload timeout in seconds.
	// If not defined or set to 0, there is no timeout.
	Timeout int

	// ReadOnly enables read only queries.
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
