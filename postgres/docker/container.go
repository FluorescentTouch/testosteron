package docker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/network"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	defaultName = "postgres"
	dbPort      = "5432/tcp"
)

type PostgresContainer struct {
	tc.Container
	ctx        context.Context
	dbName     string
	user       string
	password   string
	host       string
	port       int
	connection string
}

func (p *PostgresContainer) Host() string {
	return p.host
}

func (p *PostgresContainer) Name() string {
	return p.dbName
}

func (p *PostgresContainer) User() string {
	return p.user
}

func (p *PostgresContainer) Port() int {
	return p.port
}

func (p *PostgresContainer) Password() string {
	return p.password
}

func (p *PostgresContainer) Connection() string {
	return p.connection
}

func (p *PostgresContainer) Cleanup() error {
	return p.Terminate(context.Background())
}

func RunContainer(image string, opts ...tc.ContainerCustomizer) (*PostgresContainer, error) {
	ctx := context.Background()

	opts = append(
		[]tc.ContainerCustomizer{
			postgres.WithDatabase(defaultName),
			postgres.WithUsername(defaultName),
			postgres.WithPassword(defaultName),
			postgres.BasicWaitStrategies(),
		},
		opts...,
	)

	container, err := postgres.Run(ctx, image, opts...)
	if err != nil {
		return nil, err
	}

	pc := new(PostgresContainer)
	for i := 0; i < 3; i++ {
		pc, err = getPorts(ctx, container)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	if err != nil {
		return nil, err
	}

	cs, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	pg := &PostgresContainer{
		Container:  container,
		host:       pc.host,
		port:       pc.port,
		dbName:     pc.dbName,
		user:       pc.user,
		password:   pc.password,
		connection: cs,
	}

	return pg, nil
}

func getPorts(ctx context.Context, container tc.Container) (*PostgresContainer, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get host error: %w", err)
	}
	ports, err := container.Ports(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get port error: %w", err)
	}

	dp, err := network.ParsePort(dbPort)
	if err != nil {
		return nil, err
	}

	var hostPost string
	if len(ports[dp]) > 0 {
		hostPost = ports[dp][0].HostPort
	}

	port, err := strconv.Atoi(hostPost)
	if err != nil {
		return nil, fmt.Errorf("port '%s' parse error: %w", hostPost, err)
	}

	pg := &PostgresContainer{
		host: host,
		port: port,
	}

	return pg, nil
}
