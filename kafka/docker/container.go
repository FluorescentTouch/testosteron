package docker

import (
	"context"
	"fmt"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

type KafkaContainer struct {
	tc.Container
	brokers []string
}

func (k *KafkaContainer) Cleanup() error {
	return k.Terminate(context.Background())
}

func (k *KafkaContainer) Brokers() []string {
	return k.brokers
}

func RunContainer(image string, opts ...tc.ContainerCustomizer) (*KafkaContainer, error) {
	ctx := context.Background()

	container, err := redpanda.Run(ctx, image, opts...)
	if err != nil {
		return nil, err
	}

	seedBroker, err := container.KafkaSeedBroker(ctx)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container, brokers: []string{fmt.Sprintf("%s", seedBroker)}}, nil
}
