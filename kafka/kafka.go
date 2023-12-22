package kafka

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"github.com/FluorescentTouch/testosteron/kafka/consumer"
	"github.com/FluorescentTouch/testosteron/kafka/producer"
)

type syncProducer interface {
	io.Closer
	SendMessage(topic string, value []byte, h ...sarama.RecordHeader) error
	SendKeyMessage(topic string, key, value []byte, h ...sarama.RecordHeader) error
}

type Client struct {
	client   sarama.Client
	group    sarama.ConsumerGroup
	admin    sarama.ClusterAdmin
	producer syncProducer

	t *testing.T
}

func NewClient(t *testing.T, addr []string) *Client {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_5_1_0
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	c, err := sarama.NewClient(addr, cfg)
	if err != nil {
		t.Errorf("new kafka client error: %v", err)
		return nil
	}

	ca, err := sarama.NewClusterAdmin(addr, cfg)
	if err != nil {
		t.Errorf("new kafka cluster admin error: %v", err)
		return nil
	}

	cg, err := sarama.NewConsumerGroupFromClient(uuid.New().String(), c)
	if err != nil {
		t.Errorf("new consumer group error: %v", err)
		return nil
	}

	p, err := producer.New(c)
	if err != nil {
		t.Errorf("new producer error: %v", err)
		return nil
	}

	client := &Client{
		client:   c,
		group:    cg,
		t:        t,
		admin:    ca,
		producer: p,
	}

	t.Cleanup(func() {
		err = cg.Close()
		if err != nil {
			t.Errorf("consumer group close error: %v", err)
		}
		err = client.cleanup()
		if err != nil {
			t.Errorf("KafkaClient cleanup error: %v", err)
		}
	})

	return client
}

func (c *Client) Consume(ctx context.Context, topic string) *sarama.ConsumerMessage {
	msgChan := make(chan *sarama.ConsumerMessage)
	errChan := make(chan error)

	go func() {
		err := c.group.Consume(ctx, []string{topic}, consumer.NewChanConsumer(ctx, c.t, msgChan))
		if err != nil {
			errChan <- err
		}
	}()

	var (
		msg *sarama.ConsumerMessage
		err error
	)

	select {
	case <-ctx.Done():
		c.t.Errorf("consume to topic: '%s' stop with context.Done()", topic)
	case msg = <-msgChan:
	case err = <-errChan:
		c.t.Errorf("KafkaClient consume error: %v", err)
	}

	return msg
}

func (c *Client) cleanup() error {
	c.t.Helper()

	err := c.producer.Close()
	if err != nil {
		return fmt.Errorf("cant close producer")
	}

	// delete all topics
	topics, err := c.client.Topics()
	if err != nil {
		return fmt.Errorf("cant retrieve topic list")
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
		return fmt.Errorf("cant close client")
	}
	return nil
}

func (c *Client) Produce(topic string, value []byte, headers ...sarama.RecordHeader) {
	err := c.producer.SendMessage(topic, value, headers...)
	if err != nil {
		c.t.Errorf("KafkaClient Produce error: %v", err)
	}
}

func (c *Client) ProduceWithKey(topic string, key, value []byte, headers ...sarama.RecordHeader) {
	err := c.producer.SendKeyMessage(topic, key, value, headers...)
	if err != nil {
		c.t.Errorf("KafkaClient ProduceWithKey error: %v", err)
	}
}
