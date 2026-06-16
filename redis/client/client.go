package client

import (
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	ucl redis.UniversalClient

	t *testing.T
}

func NewClient(t *testing.T, addr string) *Client {
	c := &Client{
		t: t,
		ucl: redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs: []string{addr},
		}),
	}

	return c
}

func (c *Client) Client() redis.UniversalClient {
	return c.ucl
}

func (c *Client) cleanup() error {
	c.t.Helper()

	// close client
	err := c.ucl.Close()
	if err != nil {
		return fmt.Errorf("cant close client: %s", err)
	}
	return nil
}
