package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host        string `validate:"required"`
	User        string `validate:"required"`
	Port        int    `validate:"gt=0"`
	DbName      string `validate:"required"`
	Password    string
	SslMode     bool
	MaxOpenConn int
	MaxIdleConn int
	Wait        bool
}

func (c Config) sslMode() string {
	if c.SslMode {
		return "enable"
	}
	return "disable"
}

func (c Config) String() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DbName,
		c.sslMode(),
		c.MaxOpenConn,
	)
}

type Client struct {
	t    *testing.T
	conn *pgxpool.Pool
}

func NewClient(ctx context.Context, t *testing.T, conf Config) (*Client, error) {
	conn, err := newConnection(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("pgx new connection error: %w", err)
	}

	client := &Client{
		t:    t,
		conn: conn,
	}

	t.Cleanup(func() {
		conn.Close()
		err = client.cleanup()
		if err != nil {
			t.Errorf("PostgresClient cleanup error: %v", err)
		}
	})
	return client, nil
}

func (p *Client) cleanup() error {

	return nil
}

func newConnection(ctx context.Context, conf Config) (db *pgxpool.Pool, err error) {
	pgxConfig, err := pgxpool.ParseConfig(conf.String())
	if err != nil {
		return
	}

	db, err = pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return
	}

	return waitConnection(ctx, db)
}

func waitConnection(ctx context.Context, db *pgxpool.Pool) (*pgxpool.Pool, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		e := db.Ping(ctx)
		if e == nil {
			break
		}
		time.Sleep(time.Second)
	}

	return db, nil
}
