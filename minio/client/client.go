package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client

	t *testing.T
}

func NewClient(t *testing.T, url, username, password string) *Client {
	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(username, password, ""),
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	c := &Client{
		t:      t,
		client: minioClient,
	}

	return c
}

func (c *Client) Client() minio.Client {
	return *c.client
}

func (c *Client) cleanup() error {
	ctx := context.Background()
	buckets, err := c.client.ListBuckets(ctx)
	if err != nil {
		return err
	}

	for _, b := range buckets {
		err = c.cleanBucket(ctx, b.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) cleanBucket(ctx context.Context, name string) error {
	exists, err := c.client.BucketExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("bucket not exist %s", name)
	}

	objectCh := c.client.ListObjects(ctx, name, minio.ListObjectsOptions{
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return fmt.Errorf("obj err: %w", obj.Err)
		}

		err = c.client.RemoveObject(ctx, name, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("remove error %s: %w", obj.Key, err)
		}
	}

	return nil
}
