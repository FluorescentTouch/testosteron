package steron

import (
	"context"
	"fmt"
	"testing"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type KafkaClient struct {
	client sarama.Client
	group  sarama.ConsumerGroup
	admin  sarama.ClusterAdmin

	t *testing.T
}

func newKafkaClient(t *testing.T, addr []string) *KafkaClient {
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

	client := &KafkaClient{
		client: c,
		group:  cg,
		t:      t,
		admin:  ca,
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

func (k *KafkaClient) Consume(topic string) *sarama.ConsumerMessage {
	ctx, cancel := context.WithCancel(context.Background())

	msgChan := make(chan *sarama.ConsumerMessage)
	errChan := make(chan error)

	go func() {
		err := k.group.Consume(ctx, []string{topic}, consumer{c: msgChan, ctx: ctx, t: k.t})
		if err != nil {
			errChan <- err
		}
	}()

	var (
		msg *sarama.ConsumerMessage
		err error
	)

	select {
	case msg = <-msgChan:
	case err = <-errChan:
		k.t.Errorf("KafkaClient consume error: %v", err)
	}

	cancel()

	return msg
}

func (k *KafkaClient) cleanup() error {
	k.t.Helper()

	// delete all topics
	t, err := k.client.Topics()
	if err != nil {
		return fmt.Errorf("cant retrieve topic list")
	}

	for _, topic := range t {
		err = k.admin.DeleteTopic(topic)
		if err != nil {
			return fmt.Errorf("cant delete topic: %s", topic)
		}
	}
	// close client
	err = k.client.Close()
	if err != nil {
		return fmt.Errorf("cant close client")
	}
	return nil
}

type consumer struct {
	c   chan *sarama.ConsumerMessage
	ctx context.Context
	t   *testing.T
}

func (k consumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	close(k.c)
	return nil
}

func (k consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-k.ctx.Done():
			return nil
		case m := <-claim.Messages():
			if m == nil {
				continue
			}
			k.c <- m
			session.MarkMessage(m, k.t.Name())
		}
	}
}

func (k *KafkaClient) Produce(topic string, data []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
	err := k.produce(msg)
	if err != nil {
		k.t.Errorf("KafkaClient Produce error: %v", err)
	}
}

func (k *KafkaClient) ProduceWithKey(topic string, key, data []byte, headers ...sarama.RecordHeader) {
	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(data),
		Headers: headers,
	}
	err := k.produce(msg)
	if err != nil {
		k.t.Errorf("KafkaClient ProduceWithKey error: %v", err)
	}
}

func (k *KafkaClient) produce(msg *sarama.ProducerMessage) error {
	p, err := sarama.NewSyncProducerFromClient(k.client)
	if err != nil {
		return fmt.Errorf("new producer: %w", err)
	}
	_, _, err = p.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}
