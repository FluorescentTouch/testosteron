package reader

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type partitioner struct {
	partition sarama.PartitionConsumer
}

func newPartitioner(c sarama.Consumer, topic string, offset int64) (*partitioner, error) {
	pc, err := c.ConsumePartition(topic, 0, offset)
	if err != nil {
		err = fmt.Errorf("sarama.Consumer.ConsumePartition error: %s", err)
	}
	pr := partitioner{
		partition: pc,
	}
	return &pr, err
}

func (p *partitioner) Message(ctx context.Context) *sarama.ConsumerMessage {
	chMessages := p.partition.Messages()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-chMessages:
			return msg
		}
	}
}
