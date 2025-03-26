package simple

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	steron "github.com/FluorescentTouch/testosteron"
	"github.com/FluorescentTouch/testosteron/elastic/docker"
)

func TestMain(m *testing.M) {
	cfg, err := steron.Init(
		steron.NewKafkaService(),
		steron.NewPostgresService(),
		steron.NewElasticSearchService(docker.WithImage("elasticsearch:7.17.17")),
	)
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
