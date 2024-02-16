package steron

import (
	"testing"

	"github.com/FluorescentTouch/testosteron/docker"
	"github.com/FluorescentTouch/testosteron/kafka"
	"github.com/FluorescentTouch/testosteron/sync"
)

type KafkaHelper struct {
	clients sync.Map[KafkaClient] // t.Name:Client

	broker *docker.Kafka
}

func (h *KafkaHelper) Client(t *testing.T) KafkaClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	broker := h.broker

	// init kafka for single test if not initialized globally
	if broker == nil {
		b, err := docker.NewKafka()
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

	c := kafka.NewClient(t, broker.Brokers())
	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
