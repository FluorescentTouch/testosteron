package client

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

const postgresDriver = "postgres"

type Config struct {
	Host     string
	User     string
	Port     int
	DbName   string
	Password string
}

func (c Config) String() string {
	return fmt.Sprintf(
		"host=%s user=%s port=%d dbname=%s password=%s sslmode=disable binary_parameters=yes",
		c.Host,
		c.User,
		c.Port,
		c.DbName,
		c.Password,
	)
}

type Client struct {
	t    *testing.T
	conn *sql.DB
}

func NewClientPg(ctx context.Context, t *testing.T, conf Config) (*Client, error) {
	client := &Client{}

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
			t.Errorf("PostgresClient cleanup error: %s", err)
		}
	})
	return client, nil
}

func (c *Client) cleanup() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("postgres connect close error: %w", err)
	}
	return nil
}

func (c *Client) newConnection(ctx context.Context, conf Config) (*sql.DB, error) {
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

func (c *Client) DB() *sql.DB {
	return c.conn
}

func (c *Client) Migrate(migrateDir string) error {
	migrationsList := &migrate.FileMigrationSource{
		Dir: migrateDir,
	}

	n, err := migrate.Exec(c.conn, postgresDriver, migrationsList, migrate.Up)
	if err != nil {
		return fmt.Errorf("pg migrate error: %w", err)
	}

	c.t.Logf("Applied %d migrations. sourse: %s", n, migrationsList)
	return nil
}
