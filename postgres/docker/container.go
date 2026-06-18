package docker

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	defaultName = "postgres"
	dbPort      = "5432/tcp"
)

var connRx, _ = regexp.Compile(`^postgres:\/\/(\w+):(\w+)@(\w+):(\d+)\/(\w+)\?(.+)?$`)

type connConfig struct {
	dbName   string
	user     string
	password string
	host     string
	port     int
}

func newConnConfig(conn string) (c connConfig, err error) {
	res := connRx.FindStringSubmatch(conn)
	if len(res) != 7 {
		err = fmt.Errorf("wrong connection regex result")
	}

	port, err := strconv.Atoi(res[4])
	if err != nil {
		return
	}

	c = connConfig{
		dbName:   res[5],
		user:     res[1],
		password: res[2],
		host:     res[3],
		port:     port,
	}

	return
}

type PostgresContainer struct {
	tc.Container
	ctx        context.Context
	conf       connConfig
	connection string
}

func (p *PostgresContainer) Host() string {
	return p.conf.host
}

func (p *PostgresContainer) Name() string {
	return p.conf.dbName
}

func (p *PostgresContainer) User() string {
	return p.conf.user
}

func (p *PostgresContainer) Port() int {
	return p.conf.port
}

func (p *PostgresContainer) Password() string {
	return p.conf.password
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

	cs, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	conf, err := newConnConfig(cs)
	if err != nil {
		return nil, err
	}

	pg := &PostgresContainer{
		Container:  container,
		conf:       conf,
		connection: cs,
	}

	return pg, nil
}
