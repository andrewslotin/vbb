package vbb

import "net/http"

// Client is a VBB API client
type Client struct {
	c *http.Client
}

// New returns a new instance of VBB API client
func New(c *http.Client) *Client {
	if c == nil {
		c = http.DefaultClient
	}

	return &Client{
		c: c,
	}
}
