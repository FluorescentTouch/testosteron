package docker

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/mod/semver"
)

const publicPort = nat.Port("9093/tcp")

const dockerImage = "confluentinc/confluent-local:7.5.0"
const (
	starterScript = "/usr/sbin/testcontainers_start.sh"

	// starterScript {
	starterScriptContent = `#!/bin/bash
source /etc/confluent/docker/bash-config
export KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://%s:%d,BROKER://%s:9092
echo Starting Kafka KRaft mode
sed -i '/KAFKA_ZOOKEEPER_CONNECT/d' /etc/confluent/docker/configure
echo 'kafka-storage format --ignore-formatted -t "$(kafka-storage random-uuid)" -c /etc/kafka/kafka.properties' >> /etc/confluent/docker/configure
echo '' > /etc/confluent/docker/ensure
/etc/confluent/docker/configure
/etc/confluent/docker/launch`
	// }
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

func RunContainer(opts ...tc.ContainerCustomizer) (*KafkaContainer, error) {
	ctx := context.Background()
	req := tc.ContainerRequest{
		Image:        dockerImage,
		ExposedPorts: []string{string(publicPort)},
		Env: map[string]string{
			// envVars {
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_REST_BOOTSTRAP_SERVERS":                   "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_FLUSH_INTERVAL_MESSAGES":              fmt.Sprintf("%d", math.MaxInt),
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 fmt.Sprintf("1@%s:9094", "localhost"),
			// }
		},
		Entrypoint: []string{"sh"},
		// this CMD will wait for the starter script to be copied into the container and then execute it
		Cmd: []string{"-c", "while [ ! -f " + starterScript + " ]; do sleep 0.1; done; bash " + starterScript},
		LifecycleHooks: []tc.ContainerLifecycleHooks{
			{
				PostStarts: []tc.ContainerHook{
					// 1. copy the starter script into the container
					func(ctx context.Context, c tc.Container) error {
						host, err := c.Host(ctx)
						if err != nil {
							return err
						}

						port, err := c.MappedPort(ctx, publicPort)
						if err != nil {
							return err
						}

						scriptContent := fmt.Sprintf(starterScriptContent, host, port.Int(), host)

						return c.CopyToContainer(ctx, []byte(scriptContent), starterScript, 0o755)
					},
					// 2. wait for the Kafka server to be ready
					func(ctx context.Context, c tc.Container) error {
						return wait.ForLog(".*Transitioning from RECOVERY to RUNNING.*").AsRegexp().WaitUntilReady(ctx, c)
					},
				},
			},
		},
	}

	genericContainerReq := tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	err := validateKRaftVersion(genericContainerReq.Image)
	if err != nil {
		return nil, err
	}

	container, err := tc.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, publicPort)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container, brokers: []string{fmt.Sprintf("%s:%d", host, port.Int())}}, nil
}

func WithEnv(env map[string]string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env = env
	}
}

func WithEnvValue(key, value string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env[key] = value
	}
}

func WithImage(image string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		if len(image) == 0 {
			return
		}
		req.Image = image
	}
}

func WithExposedPorts(ports []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.ExposedPorts = ports
	}
}

func WithEntrypoint(entrypoint []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Entrypoint = entrypoint
	}
}

func WithCmd(cmd []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Cmd = cmd
	}
}

func WithLifecycleHooks(lifecycleHooks []tc.ContainerLifecycleHooks) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.LifecycleHooks = lifecycleHooks
	}
}

// validateKRaftVersion validates if the image version is compatible with KRaft mode,
// which is available since version 7.0.0.
func validateKRaftVersion(fqName string) error {
	if fqName == "" {
		return fmt.Errorf("image cannot be empty")
	}

	image := fqName[:strings.LastIndex(fqName, ":")]
	version := fqName[strings.LastIndex(fqName, ":")+1:]

	if !strings.EqualFold(image, "confluentinc/confluent-local") {
		// do not validate if the image is not the official one.
		// not raising an error here, letting the image to start and
		// eventually evaluate an error if it exists.
		return nil
	}

	// semver requires the version to start with a "v"
	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.Compare(version, "v7.4.0") < 0 { // version < v7.4.0
		return fmt.Errorf("version=%s. KRaft mode is only available since version 7.4.0", version)
	}

	return nil
}
