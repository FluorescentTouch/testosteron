package steron

import "github.com/FluorescentTouch/testosteron/sync"

var helper *Helper

func init() {
	h := &Helper{
		hyperText: &HTTPHelper{
			clients: sync.MakeSyncMap[WebClient](),
			servers: sync.MakeSyncMap[WebServer](),
		},
		kafka: &KafkaHelper{
			clients: sync.MakeSyncMap[KafkaClient](),
		},
		postgres: &PostgresHelper{
			clients: sync.MakeSyncMap[PostgresClient](),
		},
	}
	helper = h
}

type Helper struct {
	cfg Config

	hyperText *HTTPHelper
	kafka     *KafkaHelper
	postgres  *PostgresHelper
}

func (h *Helper) cleanup() {
	if h.hyperText.mainServer != nil {
		h.hyperText.mainServer.Cleanup()
	}
	if h.kafka.broker != nil {
		_ = h.kafka.broker.Cleanup()
	}
	if h.postgres.database != nil {
		_ = h.postgres.database.Cleanup()
	}
}

func (h *Helper) HTTP() *HTTPHelper {
	return h.hyperText
}

func (h *Helper) Kafka() *KafkaHelper {
	return h.kafka
}

func (h *Helper) Postgres() *PostgresHelper {
	return h.postgres
}
