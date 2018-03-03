package redis3

// Ping AWS S3 "database" bucket to check connectivity.
func (c *Client) Ping() error {
	return c.checkS3Client()
}
