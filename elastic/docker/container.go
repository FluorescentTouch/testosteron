package docker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
)

const (
	defaultPort  = "9200/tcp"
	defaultImage = "elasticsearch:6.5.4"
)

type ElasticsearchContainer struct {
	tc.Container
	ctx  context.Context
	host string
}

func (e *ElasticsearchContainer) Host() string {
	return e.host
}

func (e *ElasticsearchContainer) Cleanup() error {
	return e.Terminate(e.ctx)
}

func RunContainer(opts ...tc.ContainerCustomizer) (*ElasticsearchContainer, error) {
	ctx := context.Background()
	req := tc.ContainerRequest{
		Image: defaultImage,
		Env: map[string]string{
			"discovery.type":                     "single-node",
			"bootstrap.memory_lock":              "true",
			"network.host":                       "0.0.0.0",
			"transport.host":                     "0.0.0.0",
			"discovery.zen.minimum_master_nodes": "1",
			"xpack.license.self_generated.type":  "trial",
			"xpack.security.enabled":             "false",
			"ES_JAVA_OPTS":                       "-Xms512m -Xmx512m",
			//"xpack.monitoring.enabled":           "false",
			//"xpack.watcher.enabled":              "false",
			//"xpack.ml.enabled":                   "false",
			//"http.cors.enabled":                  "true",
			//"http.cors.allow-origin":             "*",
			//"http.cors.allow-methods":            "OPTIONS, HEAD, GET, POST, PUT, DELETE",
			//"http.cors.allow-headers":            "X-Requested-With,X-Auth-Token,Content-Type, Content-Length",
			//"logger.level":                       "debug",
		},
		ExposedPorts: []string{defaultPort},
	}

	genericContainerReq := tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := tc.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	address := ""
	for i := 0; i < 3; i++ {
		address, err = getAddress(ctx, container)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	if err != nil {
		return nil, err
	}

	es := &ElasticsearchContainer{
		Container: container,
		ctx:       ctx,
		host:      address,
	}

	return es, nil
}

func getAddress(ctx context.Context, container tc.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("postgres get host error: %w", err)
	}
	ports, err := container.Ports(ctx)
	if err != nil {
		return "", fmt.Errorf("postgres get port error: %w", err)
	}

	var hostPost string
	if len(ports[defaultPort]) > 0 {
		hostPost = ports[defaultPort][0].HostPort
	}

	port, err := strconv.Atoi(hostPost)
	if err != nil {
		return "", fmt.Errorf("port '%s' parse error: %w", hostPost, err)
	}

	return fmt.Sprintf("%s:%d", host, port), nil
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
