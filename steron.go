package steron

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/IBM/sarama"

	"github.com/FluorescentTouch/testosteron/docker"
)

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

type Config struct {
	kafkaBrokers []string
}

func (c Config) KafkaBrokers() []string {
	return c.kafkaBrokers
}

func Init(options ...option) (Config, error) {
	flag.Parse()

	for _, o := range options {
		err := o(helper)
		if err != nil {
			return Config{}, err
		}
	}
	return helper.cfg, nil
}

func Cleanup() {
	helper.cleanup()
}

func AddKafka(h *Helper) error {
	broker, err := docker.NewKafka()
	if err != nil {
		return fmt.Errorf("kafka init error: %w", err)
	}
	h.kafka.broker = broker
	h.cfg.kafkaBrokers = broker.Brokers()
	return nil
}

func AddPostgres(h *Helper) error {
	database, err := docker.NewPostgres()
	if err != nil {
		return fmt.Errorf("postgres init error: %w", err)
	}
	h.postgres.database = database
	return nil
}

func HTTP() *HTTPHelper {
	return helper.HTTP()
}

type option func(*Helper) error

func Kafka() *KafkaHelper {
	return helper.Kafka()
}

func Postgres() *PostgresHelper {
	return helper.Postgres()
}
