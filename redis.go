package steron

import (
	"fmt"
	"testing"

	"github.com/FluorescentTouch/testosteron/v2/redis/client"
	"github.com/FluorescentTouch/testosteron/v2/redis/docker"
	"github.com/FluorescentTouch/testosteron/v2/sync"
	tc "github.com/testcontainers/testcontainers-go"
)

type RedisService struct {
	opts []tc.ContainerCustomizer
}

func NewRedisService(opts ...tc.ContainerCustomizer) *RedisService {
	return &RedisService{
		opts: opts,
	}
}

func (k *RedisService) WithHelper(h *Helper) error {
	rs, err := docker.RunContainer(k.opts...)
	if err != nil {
		return fmt.Errorf("Redis init error: %w", err)
	}
	h.Redis.client = rs
	h.Redis.opts = k.opts
	h.cfg.RedisBrokers = broker.Brokers()
	return nil
}

type RedisHelper struct {
	clients sync.Map[RedisClient] // t.Name:Client
	opts    []tc.ContainerCustomizer
	redis   *docker.RedisContainer
}

func (h *RedisHelper) Client(t *testing.T) RedisClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	rs := h.redis

	// init Redis for single test if not initialized globally
	if rs == nil {
		b, err := docker.RunContainer(h.opts...)
		if err != nil {
			t.Errorf("new broker err: %s", err)
			return nil
		}

		t.Cleanup(func() {
			err = b.Cleanup()
			if err != nil {
				t.Errorf("broker cleanup err: %s", err)
			}
		})

		rs = b
	}

	c := client.NewClient(t, rs.Options().Addr)
	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
