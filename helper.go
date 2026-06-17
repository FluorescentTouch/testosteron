package steron

import "github.com/FluorescentTouch/testosteron/v2/sync"

var helper *Helper

func init() {
	h := &Helper{
		http: &HTTPHelper{
			clients: sync.MakeSyncMap[WebClient](),
			servers: sync.MakeSyncMap[WebServer](),
		},
		kafka: &KafkaHelper{
			clients: sync.MakeSyncMap[KafkaClient](),
		},
		postgres: &PostgresHelper{
			clients: sync.MakeSyncMap[DbClient](),
		},
		elasticSearch: &ElasticSearchHelper{
			clients: sync.MakeSyncMap[ElasticSearchClient](),
		},
		redis: &RedisHelper{
			clients: sync.MakeSyncMap[RedisClient](),
		},
		minio: &MinioHelper{
			clients: sync.MakeSyncMap[MinioClient](),
		},
	}
	helper = h
}

type Helper struct {
	cfg Config

	http          *HTTPHelper
	kafka         *KafkaHelper
	postgres      *PostgresHelper
	elasticSearch *ElasticSearchHelper
	redis         *RedisHelper
	minio         *MinioHelper
}

func (h *Helper) cleanup() {
	if h.http.mainServer != nil {
		h.http.mainServer.Cleanup()
	}
	if h.kafka.broker != nil {
		_ = h.kafka.broker.Cleanup()
	}
	if h.postgres.database != nil {
		_ = h.postgres.database.Cleanup()
	}
}

func (h *Helper) HTTP() *HTTPHelper {
	return h.http
}

func (h *Helper) Kafka() *KafkaHelper {
	return h.kafka
}

func (h *Helper) Postgres() *PostgresHelper {
	return h.postgres
}

func (h *Helper) ElasticSearch() *ElasticSearchHelper {
	return h.elasticSearch
}

func (h *Helper) Redis() *RedisHelper {
	return h.redis
}

func (h *Helper) Minio() *MinioHelper {
	return h.minio
}
