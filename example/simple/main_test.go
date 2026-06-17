package simple

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	steron "github.com/FluorescentTouch/testosteron/v2"
)

func TestMain(m *testing.M) {
	options := []steron.Option{
		steron.NewPostgresService("docker-hub-nexus.vkteam.ru/postgres:16.6-alpine3.20"),
		steron.NewRedisService("docker-hub-nexus.vkteam.ru/redis:7.4-alpine"),
		steron.NewKafkaService("docker-hub-nexus.vkteam.ru/redpandadata/redpanda"),
		steron.NewMinioService("docker-hub-nexus.vkteam.ru/minio/minio:RELEASE.2024-01-16T16-07-38Z"),
	}
	cfg, err := steron.Init(options...)
	if err != nil {
		panic(err)
	}

	_ = os.Setenv("KAFKA_BROKERS", strings.Join(cfg.KafkaBrokers(), ","))

	_ = os.Setenv("ELASTIC_SEARCH_URL", cfg.ElasticSearchAddress())

	_ = os.Setenv("POSTGRES_USER", cfg.PgConfig().User)
	_ = os.Setenv("POSTGRES_PASSWORD", cfg.PgConfig().Password)
	_ = os.Setenv("POSTGRES_HOST", cfg.PgConfig().Host)
	_ = os.Setenv("POSTGRES_PORT", strconv.Itoa(cfg.PgConfig().Port))
	_ = os.Setenv("POSTGRES_NAME", cfg.PgConfig().Name)

	env := []string{
		"KAFKA_BROKERS",
		"ELASTIC_SEARCH_URL",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_NAME",
	}

	fmt.Println("----- env -----")
	for _, key := range env {
		fmt.Printf("%s\n", os.Getenv(key))
	}
	fmt.Println("----- end -----")

	// run the app
	code := m.Run()

	steron.Cleanup()
	os.Exit(code)
}
