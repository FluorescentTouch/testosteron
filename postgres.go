package steron

import (
	"context"
	"github.com/FluorescentTouch/testosteron/db"
	"github.com/FluorescentTouch/testosteron/docker"
	"github.com/FluorescentTouch/testosteron/sync"
	"testing"
)

type PostgresClient interface {
	//
}

type PostgresHelper struct {
	clients  sync.Map[PostgresClient]
	database *docker.Postgres
}

func (p *PostgresHelper) Client(t *testing.T) PostgresClient {
	if c, ok := p.clients.Get(t.Name()); ok {
		return c
	}

	database := p.database
	if database == nil {
		d, err := docker.NewPostgres()
		if err != nil {
			t.Errorf("new database error: %v", err)
			return nil
		}
		t.Cleanup(func() {
			err = d.Cleanup()
			if err != nil {
				t.Errorf("database cleanup error: %v", err)
			}
		})

		database = d
	}

	conf := db.Config{
		Host:     database.Host(),
		User:     database.User(),
		Port:     0,
		DbName:   database.Name(),
		Password: database.Password(),
	}

	ctx := context.Background()
	c, err := db.NewClient(ctx, t, conf)
	if err != nil {
		t.Errorf("db new client error: %v", err)
		return nil
	}
	p.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		p.clients.Delete(t.Name())
	})

	return c
}
