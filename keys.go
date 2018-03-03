package redis3

import (
	"reflect"
	"time"
)

// Unlock TBD
func (c *Client) Unlock(key string) error {
	return c.unlockKey(key)
}

// Lock TBD
func (c *Client) Lock(key string) error {
	return c.lockKey(key)
}

// Del TBD
func (c *Client) Del(key string) error {
	return c.deleteKey(key)
}

// Get TBD
func (c *Client) Get(key string, value interface{}) (KeyMetadata, error) {
	metadata, err := c.downloadKey(key, value)

	return metadata, err
}

// Set TBD
func (c *Client) Set(key string, value interface{}, expiration int64) error {
	var expireTime int64

	now := time.Now()
	if expiration > 0 {
		expireTime = now.Unix() + expiration
	}

	metadata := KeyMetadata{
		ValueType:  reflect.TypeOf(value).String(),
		ExpireTime: expireTime,
		LastUpdate: now.Unix(),
	}

	return c.uploadKey(key, value, metadata)
}
