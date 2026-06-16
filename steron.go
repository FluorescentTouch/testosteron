package steron

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/redis/go-redis/v9"
)

type WebServer interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
	Addr() string
	Cleanup()
	NewDebugHandler(handlerFunc http.HandlerFunc) func(...string)
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
	CreateTopic(name string, detail *sarama.TopicDetail, validateOnly bool)
}

type RedisClient interface {
	Client() redis.UniversalClient
}

type ElasticSearchClient interface {
	Client() *es.Client
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
	esHost         string
}

func (c Config) KafkaBrokers() []string {
	return c.kafkaBrokers
}

func (c Config) PgConfig() DbConfig {
	return c.postgresConfig
}

func (c Config) ElasticSearchAddress() string {
	return c.esHost
}

func Init(options ...Option) (Config, error) {
	flag.Parse()

	for _, o := range options {
		err := o.WithHelper(helper)
		if err != nil {
			return Config{}, err
		}
	}
	return helper.cfg, nil
}

func Cleanup() {
	helper.cleanup()
}

func HTTP() *HTTPHelper {
	return helper.HTTP()
}

type Option interface {
	WithHelper(*Helper) error
}

func Kafka() *KafkaHelper {
	return helper.Kafka()
}

func Postgres() *PostgresHelper {
	return helper.Postgres()
}

func ElasticSearch() *ElasticSearchHelper {
	return helper.ElasticSearch()
}
