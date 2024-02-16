package steron

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"time"

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
	Consume(ctx context.Context, timeout time.Duration, topic string) *sarama.ConsumerMessage
	Produce(topic string, value []byte, h ...sarama.RecordHeader)
	ProduceWithKey(topic string, key []byte, data []byte, headers ...sarama.RecordHeader)
}

type DbClient interface {
	DB() *sql.DB
	Migrate(migrateDir string) error
}

type DbConfig struct {
	Host     string
	Name     string
	User     string
	Port     int
	Password string
}

type Config struct {
	postgresConfig DbConfig
	kafkaBrokers   []string
}

func (c Config) KafkaBrokers() []string {
	return c.kafkaBrokers
}

func (c Config) PgConfig() DbConfig {
	return c.postgresConfig
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
	h.cfg.postgresConfig = DbConfig{
		Host:     database.Host(),
		Name:     database.Name(),
		User:     database.User(),
		Port:     database.Port(),
		Password: database.Password(),
	}
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
