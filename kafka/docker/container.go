package docker

import (
	"context"
	"fmt"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

const dockerImage = "docker.redpanda.com/redpandadata/redpanda:v25.2.4"

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

func RunContainer(opts ...tc.ContainerCustomizer) (*KafkaContainer, error) {
	ctx := context.Background()

	container, err := redpanda.Run(ctx, dockerImage, opts...)
	if err != nil {
		return nil, err
	}

	seedBroker, err := container.KafkaSeedBroker(ctx)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container, brokers: []string{fmt.Sprintf("%s", seedBroker)}}, nil
}
