package docker

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

type Kafka struct {
	brokers []string

	container testcontainers.Container
}

func NewKafka() (*Kafka, error) {
	ctx := context.Background()

	kafkaC, err := kafka.RunContainer(ctx,
		kafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not start container: %w", err)
	}

	brokers, err := kafkaC.Brokers(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get brokers: %w", err)
	}

	return &Kafka{
		container: kafkaC,
		brokers:   brokers,
	}, nil
}

func (k Kafka) Brokers() []string {
	return k.brokers
}

func (k Kafka) Cleanup() error {
	ctx := context.Background()
	return k.container.Terminate(ctx)
}
