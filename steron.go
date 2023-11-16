package steron

import (
	"flag"
	"fmt"
	"testing"

	"github.com/FluorescentTouch/testosteron/docker"
	"github.com/FluorescentTouch/testosteron/sync"
)

var (
	h *Helper
)

func init() {
	helper := &Helper{
		hyperText: &HTTPHelper{
			clients: sync.MakeSyncMap[*HTTPClient](),
			servers: sync.MakeSyncMap[*HTTPServer](),
		},
		kafka: &KafkaHelper{
			clients: sync.MakeSyncMap[*KafkaClient](),
		},
	}
	h = helper
}

type Helper struct {
	cfg Config

	hyperText *HTTPHelper
	kafka     *KafkaHelper
}

type Config struct {
	KafkaBrokers []string
}

type HTTPHelper struct {
	clients sync.Map[*HTTPClient] // t.Name:Client
	servers sync.Map[*HTTPServer] // t.Name:Server

	mainServer *HTTPMainServer // server started for main init
}

type KafkaHelper struct {
	clients sync.Map[*KafkaClient] // t.Name:Client

	broker *docker.Kafka
}

func Init(options ...option) (Config, error) {
	flag.Parse()

	for _, o := range options {
		err := o(h)
		if err != nil {
			return Config{}, err
		}
	}
	return h.cfg, nil
}

func Cleanup() {
	h.cleanup()
}

func AddKafka(h *Helper) error {
	broker, err := docker.NewKafka()
	if err != nil {
		return fmt.Errorf("kafka init error: %v", err)
	}
	h.kafka.broker = broker
	return nil
}

func (h *Helper) cleanup() {
	if h.hyperText.mainServer != nil {
		h.hyperText.mainServer.cleanup()
	}
	if h.kafka.broker != nil {
		_ = h.kafka.broker.Cleanup()
	}
}

func (h *Helper) HTTP() *HTTPHelper {
	return h.hyperText
}

func HTTP() *HTTPHelper {
	return h.HTTP()
}

type option func(*Helper) error

func (h *HTTPHelper) Client(t *testing.T) *HTTPClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	c := newHTTPClient(t)

	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}

func (h *HTTPHelper) Server(t *testing.T) *HTTPServer {
	if s, ok := h.servers.Get(t.Name()); ok {
		return s
	}

	s := newHTTPServer(t)

	h.servers.Set(t.Name(), s)

	t.Cleanup(func() {
		h.servers.Delete(t.Name())
	})

	return s
}

func (h *HTTPHelper) ServerMain(m *testing.M) *HTTPMainServer {
	h.mainServer = newHTTPMainServer(m)
	return h.mainServer
}

func Kafka() *KafkaHelper {
	return h.Kafka()
}

func (h *Helper) Kafka() *KafkaHelper {
	return h.kafka
}

func (h *KafkaHelper) Client(t *testing.T) *KafkaClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	broker := h.broker

	// init kafka for single test if not initialized globally
	if broker == nil {
		b, err := docker.NewKafka()
		if err != nil {
			t.Errorf("new broker err: %v", err)
			return nil
		}

		t.Cleanup(func() {
			err := b.Cleanup()
			t.Errorf("broker cleanup err: %v", err)
		})

		broker = b
	}

	c := newKafkaClient(t, broker.Brokers())

	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
