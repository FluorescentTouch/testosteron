package steron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// HTTPClient emulates any incoming http requests.
type HTTPClient struct {
	c http.Client
	t *testing.T
}

func newHTTPClient(t *testing.T) *HTTPClient {
	t.Helper()

	c := &HTTPClient{c: http.Client{}, t: t}
	t.Cleanup(func() {
		c.cleanup()
	})
	return c
}

func (c *HTTPClient) cleanup() {}

func (c *HTTPClient) do(method, url string, bodyIn []byte) (*http.Response, error) {
	c.t.Helper()

	var b io.Reader
	if bodyIn != nil {
		b = bytes.NewBuffer(bodyIn)
	}
	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, fmt.Errorf("NewRequest error: %w", err)
	}

	return c.c.Do(req)
}

// Do allows any custom requests.
func (c *HTTPClient) Do(req *http.Request) *http.Response {
	c.t.Helper()

	resp, err := c.c.Do(req)
	if err != nil {
		c.t.Errorf("client Do error: %v", err)
	}
	return resp
}

// Get makes GET requests and returns *http.Response.
func (c *HTTPClient) Get(url string) *http.Response {
	c.t.Helper()

	resp, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		c.t.Errorf("HTTPClient Get: %v", err)
		return nil
	}
	return resp
}

// GetJSON makes GET request and parse response to provided destination.
// Fails if destination provided is invalid.
func (c *HTTPClient) GetJSON(url string, dst any) {
	c.t.Helper()

	resp, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		c.t.Errorf("HTTPClient Get: %v", err)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(dst)
	if err != nil {
		c.t.Errorf("HTTPClient Decode: %v", err)
		return
	}
}
