package steron

import (
	"context"
	"fmt"
	"testing"

	tc "github.com/testcontainers/testcontainers-go"

	"github.com/FluorescentTouch/testosteron/v2/postgres/client"
	"github.com/FluorescentTouch/testosteron/v2/postgres/docker"
	"github.com/FluorescentTouch/testosteron/v2/sync"
)

type PostgresService struct {
	opts []tc.ContainerCustomizer
}

func NewPostgresService(opts ...tc.ContainerCustomizer) *PostgresService {
	return &PostgresService{
		opts: opts,
	}
}

func (p *PostgresService) WithHelper(h *Helper) error {
	database, err := docker.RunContainer(p.opts...)
	if err != nil {
		return fmt.Errorf("postgres init error: %w", err)
	}
	h.postgres.database = database
	h.postgres.opts = p.opts
	h.cfg.postgresConfig = DbConfig{
		Host:     database.Host(),
		Name:     database.Name(),
		User:     database.User(),
		Port:     database.Port(),
		Password: database.Password(),
	}
	return nil
}

type PostgresHelper struct {
	clients  sync.Map[DbClient]
	opts     []tc.ContainerCustomizer
	database *docker.PostgresContainer
}

func (p *PostgresHelper) Client(t *testing.T) DbClient {
	if c, ok := p.clients.Get(t.Name()); ok {
		return c
	}

	database := p.database
	if database == nil {
		d, err := docker.RunContainer(p.opts...)
		if err != nil {
			t.Errorf("new database error: %s", err)
			return nil
		}
		t.Cleanup(func() {
			err = d.Cleanup()
			if err != nil {
				t.Errorf("database cleanup error: %s", err)
			}
		})

		database = d
	}

	conf := client.Config{
		Host:     database.Host(),
		User:     database.User(),
		Port:     database.Port(),
		DbName:   database.Name(),
		Password: database.Password(),
	}

	ctx := context.Background()
	c, err := client.NewClientPg(ctx, t, conf)
	if err != nil {
		t.Errorf("db new client error: %s", err)
		return nil
	}
	p.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		p.clients.Delete(t.Name())
	})

	return c
}
