package steron

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"testing"

	"github.com/IBM/sarama"

	"github.com/FluorescentTouch/testosteron/docker"
	"github.com/FluorescentTouch/testosteron/http/client"
	"github.com/FluorescentTouch/testosteron/http/server"
	"github.com/FluorescentTouch/testosteron/kafka"
	"github.com/FluorescentTouch/testosteron/sync"
)

var (
	h *Helper
)

func init() {
	helper := &Helper{
		hyperText: &HTTPHelper{
			clients: sync.MakeSyncMap[WebClient](),
			servers: sync.MakeSyncMap[WebServer](),
		},
		kafka: &KafkaHelper{
			clients: sync.MakeSyncMap[KafkaClient](),
		},
	}
	h = helper
}

type WebServer interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
	Addr() string
	Cleanup()
}

type WebClient interface {
	Do(req *http.Request) *http.Response
	Get(url string) *http.Response
	GetJSON(url string, dst any)
}

type KafkaClient interface {
	Consume(ctx context.Context, topic string) *sarama.ConsumerMessage
	Produce(topic string, value []byte, h ...sarama.RecordHeader)
	ProduceWithKey(topic string, key []byte, data []byte, headers ...sarama.RecordHeader)
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
	clients sync.Map[WebClient] // t.Name:Client
	servers sync.Map[WebServer] // t.Name:Server

	mainServer WebServer // server started for main init
}

type KafkaHelper struct {
	clients sync.Map[KafkaClient] // t.Name:Client

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
		h.hyperText.mainServer.Cleanup()
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

func (h *HTTPHelper) Client(t *testing.T) WebClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	c := client.NewHTTPClient(t)

	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}

func (h *HTTPHelper) Server(t *testing.T) WebServer {
	if s, ok := h.servers.Get(t.Name()); ok {
		return s
	}

	s := server.NewHTTPServer(t)

	h.servers.Set(t.Name(), s)

	t.Cleanup(func() {
		h.servers.Delete(t.Name())
	})

	return s
}

func (h *HTTPHelper) ServerMain(m *testing.M) WebServer {
	h.mainServer = server.NewHTTPMainServer(m)
	return h.mainServer
}

func Kafka() *KafkaHelper {
	return h.Kafka()
}

func (h *Helper) Kafka() *KafkaHelper {
	return h.kafka
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
			t.Errorf("new broker err: %v", err)
			return nil
		}

		t.Cleanup(func() {
			err := b.Cleanup()
			t.Errorf("broker cleanup err: %v", err)
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
