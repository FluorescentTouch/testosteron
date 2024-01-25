package docker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	dbUserName = "db_user"
	dbPassword = "db_password"
	dbName     = "testosterone"
	dbPort     = "5432/tcp"
)

type Postgres struct {
	ctx       context.Context
	host      string
	port      int
	user      string
	password  string
	name      string
	container testcontainers.Container
}

func NewPostgres() (*Postgres, error) {
	ctx := context.Background()

	options := []testcontainers.ContainerCustomizer{
		postgres.WithUsername(dbUserName),
		postgres.WithPassword(dbPassword),
		postgres.WithDatabase(dbName),
	}
	container, err := postgres.RunContainer(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("could not start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get host error: %w", err)
	}
	ports, err := container.Ports(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get port error: %w", err)
	}

	var hostPost string
	if len(ports[dbPort]) > 0 {
		hostPost = ports[dbPort][0].HostPort
	}

	port, err := strconv.Atoi(hostPost)
	if err != nil {
		return nil, fmt.Errorf("port '%s' parse error: %w", hostPost, err)
	}
	ps := &Postgres{
		ctx:       context.Background(),
		host:      host,
		port:      port,
		user:      dbUserName,
		password:  dbPassword,
		name:      dbName,
		container: container,
	}
	return ps, nil
}

func (p *Postgres) Host() string {
	return p.host
}

func (p *Postgres) Name() string {
	return p.name
}

func (p *Postgres) User() string {
	return p.user
}

func (p *Postgres) Port() int {
	return p.port
}

func (p *Postgres) Password() string {
	return p.password
}

func (p *Postgres) Cleanup() error {
	return p.container.Terminate(p.ctx)
}
