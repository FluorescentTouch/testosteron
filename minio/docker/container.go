package docker

import (
	"context"
	"errors"

	tc "github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

type Config struct {
	Url      string
	Username string
	Password string
}
type MinioContainer struct {
	tc.Container

	cfg Config
}

func (c *MinioContainer) Cleanup() error {
	return c.Terminate(context.Background())
}

func RunContainer(image string, opts ...tc.ContainerCustomizer) (*MinioContainer, error) {
	ctx := context.Background()
	opts = append(
		[]tc.ContainerCustomizer{
			tcminio.WithUsername("admin"),
			tcminio.WithPassword("admin"),
		},
		opts...,
	)

	container, err := tcminio.Run(ctx,
		image,
		tcminio.WithUsername("thisismyuser"), tcminio.WithPassword("thisismypassword"))

	if container == nil {
		return nil, errors.New("container is nil")
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &MinioContainer{Container: container, cfg: Config{Url: uri, Username: container.Username, Password: container.Password}}, nil
}

func (c *MinioContainer) Config() Config {
	return c.cfg
}
