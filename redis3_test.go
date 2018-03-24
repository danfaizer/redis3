package redis3_test

import (
	"testing"
	"time"

	"github.com/danfaizer/redis3"
)

func TestRediS3ClientNoBucket(t *testing.T) {
	var err error
	_, err = redis3.NewClient(
		&redis3.Options{})
	if err.Error() != "client: no bucket specified in options" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestRediS3ClientNoRegion(t *testing.T) {
	var err error
	_, err = redis3.NewClient(
		&redis3.Options{
			Bucket: "redis3-test",
		})
	if err.Error() != "client: no region specified in options" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestRediS3ClientUnexsistingBucket(t *testing.T) {
	var err error
	_, err = redis3.NewClient(
		&redis3.Options{
			Bucket:   "redis3-test-wrong",
			Region:   "eu-west-1",
			Endpoint: "http://127.0.0.1:5001",
		})
	if err.Error() != "client: specified bucket does not exist" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestRediS3Client(t *testing.T) {
	client, err := redis3.NewClient(
		&redis3.Options{
			Bucket:             "redis3-test",
			AutoCreateBucket:   true,
			Region:             "eu-west-1",
			Timeout:            1,
			Endpoint:           "http://127.0.0.1:5001",
			EnforceConsistency: true,
		})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// Test Ping
	err = client.Ping()
	if err != nil {
		t.Errorf("ping failed with error: %s", err)
	}

	// Test Get unexisting key
	var unexistingValue string
	_, err = client.Get("unexistingKey", &unexistingValue)
	if err.Error() != "key not found" {
		t.Errorf("unexpected error: %s", err)
	}

	// Test Set key1
	err = client.Set("key1", "key1 value", 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Sleep so key1 expires
	time.Sleep(2 * time.Second)
	var key1 string
	_, err = client.Get("key1", &key1)
	if err.Error() != "key not found" {
		t.Errorf("unexpected error: %s", err)
	}

	// Test Set key2
	err = client.Set("key2", "key2 value", 0)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	var key2 string
	_, err = client.Get("key2", &key2)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if key2 != "key2 value" {
		t.Errorf("unexpected value for key2: %s", key2)
	}
}
