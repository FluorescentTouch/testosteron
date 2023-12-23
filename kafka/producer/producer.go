package producer

import (
	"fmt"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func New(client sarama.Client) (*Producer, error) {
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}
	p := &Producer{
		producer: syncProducer,
	}

	return p, nil
}

func (p *Producer) SendMessage(topic string, value []byte, h ...sarama.RecordHeader) error {
	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Value:   sarama.ByteEncoder(value),
		Headers: h,
	}

	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *Producer) SendKeyMessage(topic string, key, value []byte, h ...sarama.RecordHeader) error {
	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: h,
	}

	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
