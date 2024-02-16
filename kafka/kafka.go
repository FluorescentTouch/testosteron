package kafka

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/IBM/sarama"

	"github.com/FluorescentTouch/testosteron/kafka/producer"
	"github.com/FluorescentTouch/testosteron/kafka/reader"
)

type syncProducer interface {
	io.Closer
	SendMessage(topic string, value []byte, h ...sarama.RecordHeader) error
	SendKeyMessage(topic string, key, value []byte, h ...sarama.RecordHeader) error
}

type Client struct {
	client   sarama.Client
	consumer sarama.Consumer
	admin    sarama.ClusterAdmin
	producer syncProducer
	reader   *reader.Reader

	t *testing.T
}

func NewClient(t *testing.T, addr []string) *Client {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_5_1_0
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	client := &Client{
		t: t,
	}
	var err error

	client.client, err = sarama.NewClient(addr, cfg)
	if err != nil {
		t.Errorf("new kafka client error: %s", err)
		return nil
	}

	client.admin, err = sarama.NewClusterAdmin(addr, cfg)
	if err != nil {
		t.Errorf("new kafka cluster admin error: %s", err)
		return nil
	}

	client.consumer, err = sarama.NewConsumerFromClient(client.client)
	if err != nil {
		t.Errorf("new consumer group error: %s", err)
		return nil
	}

	client.producer, err = producer.New(client.client)
	if err != nil {
		t.Errorf("new producer error: %s", err)
		return nil
	}

	client.reader = reader.New(t, client.consumer)

	t.Cleanup(func() {
		err = client.cleanup()
		if err != nil {
			t.Errorf("KafkaClient cleanup error: %s", err)
		}
	})

	return client
}

func (c *Client) Consume(ctx context.Context, timeout time.Duration, topic string) *sarama.ConsumerMessage {
	return c.reader.Read(ctx, timeout, topic)
}

func (c *Client) cleanup() error {
	c.t.Helper()

	err := c.reader.Stop()
	if err != nil {
		return fmt.Errorf("cant stop kafka reader: %s", err)
	}
	err = c.producer.Close()
	if err != nil {
		return fmt.Errorf("cant close producer: %s", err)
	}

	// delete all topics
	topics, err := c.client.Topics()
	if err != nil {
		return fmt.Errorf("cant retrieve topic list: %s", err)
	}

	for _, topic := range topics {
		err = c.admin.DeleteTopic(topic)
		if err != nil {
			return fmt.Errorf("cant delete topic: %s", topic)
		}
	}
	// close client
	err = c.client.Close()
	if err != nil {
		return fmt.Errorf("cant close client: %s", err)
	}
	return nil
}

func (c *Client) Produce(topic string, value []byte, headers ...sarama.RecordHeader) {
	err := c.producer.SendMessage(topic, value, headers...)
	if err != nil {
		c.t.Errorf("KafkaClient Produce error: %s", err)
	}
}

func (c *Client) ProduceWithKey(topic string, key, value []byte, headers ...sarama.RecordHeader) {
	err := c.producer.SendKeyMessage(topic, key, value, headers...)
	if err != nil {
		c.t.Errorf("KafkaClient ProduceWithKey error: %s", err)
	}
}
