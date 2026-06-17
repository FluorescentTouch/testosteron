package steron

import (
	"fmt"
	"testing"

	"github.com/FluorescentTouch/testosteron/v2/minio/client"
	"github.com/FluorescentTouch/testosteron/v2/minio/docker"
	"github.com/FluorescentTouch/testosteron/v2/sync"
	tc "github.com/testcontainers/testcontainers-go"
)

type MinioService struct {
	opts  []tc.ContainerCustomizer
	image string
}

func NewMinioService(image string, opts ...tc.ContainerCustomizer) *MinioService {
	return &MinioService{
		opts:  opts,
		image: image,
	}
}

func (k *MinioService) WithHelper(h *Helper) error {
	mn, err := docker.RunContainer(k.image, k.opts...)
	if err != nil {
		return fmt.Errorf("minio init error: %w", err)
	}
	h.minio.minio = mn
	h.minio.image = k.image
	h.cfg.minioConfig = MinioConfig{
		Host:     mn.Config().Url,
		User:     mn.Config().Username,
		Password: mn.Config().Password,
	}
	return nil
}

type MinioHelper struct {
	clients sync.Map[MinioClient] // t.Name:Client
	opts    []tc.ContainerCustomizer
	minio   *docker.MinioContainer
	image   string
}

func (h *MinioHelper) Client(t *testing.T) MinioClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	mn := h.minio

	// init Redis for single test if not initialized globally
	if mn == nil {
		b, err := docker.RunContainer(h.image, h.opts...)
		if err != nil {
			t.Errorf("new minio err: %s", err)
			return nil
		}

		t.Cleanup(func() {
			err = b.Cleanup()
			if err != nil {
				t.Errorf("minio cleanup err: %s", err)
			}
		})

		mn = b
	}

	cfg := mn.Config()
	c := client.NewClient(t, cfg.Url, cfg.Username, cfg.Password)
	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
