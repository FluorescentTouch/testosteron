# testosteron is package for easi integrational service test with live environment

## How to use

### Add package initialization to Test_Main, or create basic Test_Main if not exists

```golang
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	steron "github.com/FluorescentTouch/testosteron"
)

func TestMain(m *testing.M) {
	// init package with required options
	cfg, err := steron.Init(steron.AddKafka)
	if err != nil {
		panic(err)
	}

	// add http server, if application requires requests for startup
	srv := steron.HTTP().ServerMain(m)
	srv.HandleFunc("/web/config", func(w http.ResponseWriter, _ *http.Request, ) {
		remoteConfig := fmt.Sprintf(`{"kafka_brokers":[%s]}`, strings.Join(cfg.KafkaBrokers(), ","))
		_, _ = w.Write([]byte(remoteConfig))
	})

	// provide test tools configuration to application
	_ = os.Setenv("REMOTE_SERVER_ADDR", srv.Addr())
	_ = os.Setenv("KAFKA_BROKERS", strings.Join(cfg.KafkaBrokers(), ","))

	// run the app
	code := m.Run()

	steron.Cleanup()
	os.Exit(code)
}
```

### Use required tools while testing application

- HTTP Server and Client to test application endpoints

```golang
func TestHTTPClientDo(t *testing.T) {
	// can be your server
	srv := steron.HTTP().Server(t)
	srv.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client := steron.HTTP().Client(t)

	req, err := http.NewRequest(http.MethodGet, srv.Addr(), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatal("status code is not 200")
	}
}
```

- You may also use helpers to make it easier
```golang
func TestHTTPClientGetJSON(t *testing.T) {
	// can be your server
	srv := steron.HTTP().Server(t)
	srv.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"value"}`))
	})
	client := steron.HTTP().Client(t)

	data := make(map[string]any, 0)

	client.GetJSON(srv.Addr(), &data)
	if len(data) == 0 {
		t.Fatal("zero values received")
	}
}
```
- Kafka Consume/Produce
```golang
func TestKafka(t *testing.T) {
	kafkaClient := steron.Kafka().Client(t)

	kafkaClient.Produce("sample_topic", []byte("msg"))
	
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	
	timeout := time.Second * 10
	msg := kafkaClient.Consume(ctx, timeout, "sample_topic")
	if len(msg.Value) == 0 {
		t.Fatal("zero len message received")
	}
}
```