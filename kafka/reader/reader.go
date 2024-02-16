package reader

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"

	"github.com/FluorescentTouch/testosteron/sync"
)

type Message struct {
	Topic     string            `json:"topic"`
	Partition int32             `json:"partition"`
	Offset    int64             `json:"offset"`
	Value     []byte            `json:"value"`
	Key       []byte            `json:"key,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type Reader struct {
	testing  *testing.T
	consumer sarama.Consumer
	offsets  sync.Map[int64]
}

func New(t *testing.T, c sarama.Consumer) *Reader {
	r := &Reader{
		testing:  t,
		consumer: c,
		offsets:  sync.MakeSyncMap[int64](),
	}
	return r
}

func (r *Reader) Stop() error {
	return r.consumer.Close()
}

func (r *Reader) Read(ctx context.Context, timeout time.Duration, topic string) *sarama.ConsumerMessage {
	if timeout == 0 {
		timeout = time.Second
	}
	currentOffset, _ := r.offsets.Get(topic)
	partitionConn, err := newPartitioner(r.consumer, topic, currentOffset)
	if err != nil {
		r.testing.Fatalf("kafka reader newPartitioner error: %s", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer func() {
		_ = partitionConn.partition.Close()
		cancel()
	}()
	message := partitionConn.Message(ctx)
	r.offsets.Set(topic, message.Offset+1)
	return message
}
