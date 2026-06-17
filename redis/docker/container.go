package docker

import (
	"context"

	r9 "github.com/redis/go-redis/v9"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

type RedisContainer struct {
	tc.Container

	options r9.Options
}

func (c *RedisContainer) Cleanup() error {
	return c.Terminate(context.Background())
}

func RunContainer(image string, opts ...tc.ContainerCustomizer) (*RedisContainer, error) {
	ctx := context.Background()

	container, err := redis.Run(ctx, image, opts...)
	if err != nil {
		return nil, err
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	options, err := r9.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	return &RedisContainer{Container: container, options: *options}, nil
}

func (c *RedisContainer) Options() r9.Options {
	return c.options
}
