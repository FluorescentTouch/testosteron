package steron

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/olivere/randport"
)

// SetEnv - use "embed" package to receive env file content
func SetEnv(content []byte) error {
	envContent := bytes.NewReader(content)
	scanner := bufio.NewScanner(envContent)

	for scanner.Scan() {
		variable := scanner.Text()

		// skip empty
		if len(variable) == 0 {
			continue
		}

		// skip comment
		if strings.HasPrefix(variable, "#") {
			continue
		}

		line := scanner.Text()
		envVar := strings.Split(line, "=")
		err := os.Setenv(envVar[0], envVar[1])
		if err != nil {
			return err
		}
	}

	err := scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

func JoinHostFromRandomPort(urlPath, env string) string {
	return fmt.Sprintf("http://localhost:%s%s", os.Getenv(env), urlPath)
}

func RandomPortEnv() string {
	return fmt.Sprintf("%d", randport.Get())
}

func Unmarshal(t *testing.T, data []byte, source any) {
	err := json.Unmarshal(data, source)
	if err != nil {
		t.Fatalf("unmarshal err to [%T]: %v", err, source)
	}
}

func ApplyMigrations(t *testing.T, migrateDirPath, scriptBefore string) {
	client := PgxClient(t)

	SqlExecFromFile(t, client.DB(), scriptBefore)
	err := client.Migrate(migrateDirPath)
	if err != nil {
		t.Fatalf("client migrage error: %v", err)
	}
}

func SqlExecFromFile(t *testing.T, conn *sql.DB, scriptSql string) {
	_, err := conn.Exec(scriptSql)
	if err != nil {
		t.Fatalf("db exec from file error: %v", err)
	}
}

func PgxClient(t *testing.T) DbClient {
	db := Postgres()
	if db == nil {
		t.Fatal("db is empty")
	}

	client := db.Client(t)
	if client == nil {
		t.Fatal("db client is empty")
	}

	return client
}

func NewKafkaClient(t *testing.T) KafkaClient {
	kc := Kafka()
	if kc == nil {
		t.Fatalf("kafka helper is empty")
	}

	client := kc.Client(t)
	if client == nil {
		t.Fatalf("kafka client is empty")
	}

	return client
}

func EnvDebugHandler(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		for _, env := range os.Environ() {
			if !strings.HasPrefix(env, prefix) {
				continue
			}

			_, _ = w.Write([]byte(fmt.Sprintf("%s\n", env)))
		}
	}
}
