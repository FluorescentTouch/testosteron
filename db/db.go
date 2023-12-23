package db

import (
	"fmt"
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
