package consumer

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
)

type ChanConsumer struct {
	t      *testing.T
	ctx    context.Context
	msgOut chan<- *sarama.ConsumerMessage
}

func NewChanConsumer(ctx context.Context, t *testing.T, m chan<- *sarama.ConsumerMessage) *ChanConsumer {
	return &ChanConsumer{
		t:      t,
		ctx:    ctx,
		msgOut: m,
	}
}

func (c *ChanConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ChanConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	close(c.msgOut)
	return nil
}

func (c *ChanConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case m := <-claim.Messages():
			if m == nil {
				continue
			}
			c.msgOut <- m
			session.MarkMessage(m, c.t.Name())
		}
	}
}
