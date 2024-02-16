package steron

import (
	"context"
	"testing"

	"github.com/FluorescentTouch/testosteron/db"
	"github.com/FluorescentTouch/testosteron/docker"
	"github.com/FluorescentTouch/testosteron/sync"
)

type PostgresHelper struct {
	clients  sync.Map[DbClient]
	database *docker.Postgres
}

func (p *PostgresHelper) Client(t *testing.T) DbClient {
	if c, ok := p.clients.Get(t.Name()); ok {
		return c
	}

	database := p.database
	if database == nil {
		d, err := docker.NewPostgres()
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

	conf := db.Config{
		Host:     database.Host(),
		User:     database.User(),
		Port:     database.Port(),
		DbName:   database.Name(),
		Password: database.Password(),
	}

	ctx := context.Background()
	c, err := db.NewClientPg(ctx, t, conf)
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
