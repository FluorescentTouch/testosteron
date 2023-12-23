package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

type ClientPg struct {
	t    *testing.T
	conn *sql.DB
}

func NewClientPg(ctx context.Context, t *testing.T, conf Config) (*ClientPg, error) {
	client := &ClientPg{}

	conn, err := client.newConnection(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("pgx new connection error: %w", err)
	}

	t.Log("successful connection to postgres database")
	client.t = t
	client.conn = conn

	t.Cleanup(func() {
		err = client.cleanup()
		if err != nil {
			t.Errorf("PostgresClient cleanup error: %v", err)
		}
	})
	return client, nil
}

func (p *ClientPg) cleanup() error {
	err := p.conn.Close()
	if err != nil {
		return fmt.Errorf("postgres connect close error: %w", err)
	}
	return nil
}

func (p *ClientPg) newConnection(ctx context.Context, conf Config) (*sql.DB, error) {
	conn, err := sqlx.Open(postgresDriver, conf.String())
	if err != nil {
		return nil, fmt.Errorf("sql open error:; %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		e := conn.DB.Ping()
		if e == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return conn.DB, nil
}

func (p *ClientPg) DB() *sql.DB {
	return p.conn
}

func (p *ClientPg) Migrate(migrateDir string) error {
	migrationsList := &migrate.FileMigrationSource{
		Dir: migrateDir,
	}

	n, err := migrate.Exec(p.conn, postgresDriver, migrationsList, migrate.Up)
	if err != nil {
		return fmt.Errorf("pg migrate error: %w", err)
	}

	p.t.Log(fmt.Sprintf("Applied %d migrations. sourse: %v", n, migrationsList))
	return nil
}
