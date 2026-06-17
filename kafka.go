package steron

import (
	"fmt"
	"testing"

	"github.com/FluorescentTouch/testosteron/v2/kafka/client"
	"github.com/FluorescentTouch/testosteron/v2/kafka/docker"
	"github.com/FluorescentTouch/testosteron/v2/sync"
	tc "github.com/testcontainers/testcontainers-go"
)

type KafkaService struct {
	opts  []tc.ContainerCustomizer
	image string
}

func NewKafkaService(image string, opts ...tc.ContainerCustomizer) *KafkaService {
	return &KafkaService{
		opts:  opts,
		image: image,
	}
}

func (k *KafkaService) WithHelper(h *Helper) error {
	broker, err := docker.RunContainer(k.image, k.opts...)
	if err != nil {
		return fmt.Errorf("kafka init error: %w", err)
	}
	h.kafka.broker = broker
	h.kafka.opts = k.opts
	h.cfg.kafkaBrokers = broker.Brokers()
	h.kafka.image = k.image
	return nil
}

type KafkaHelper struct {
	clients sync.Map[KafkaClient] // t.Name:Client
	opts    []tc.ContainerCustomizer
	broker  *docker.KafkaContainer
	image   string
}

func (h *KafkaHelper) Client(t *testing.T) KafkaClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	broker := h.broker

	// init kafka for single test if not initialized globally
	if broker == nil {
		b, err := docker.RunContainer(h.image, h.opts...)
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

		broker = b
	}

	c := client.NewClient(t, broker.Brokers())
	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
