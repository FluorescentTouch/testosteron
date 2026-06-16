package simple

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	steron "github.com/FluorescentTouch/testosteron/v2"
	"github.com/IBM/sarama"
)

func kafkaClient(t *testing.T) steron.KafkaClient {
	kc := steron.Kafka()
	if kc == nil {
		t.Fatalf("kafka helper is empty")
	}

	client := kc.Client(t)
	if client == nil {
		t.Fatalf("kafka client is empty")
	}

	return client
}

func TestCheckEnv(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kc := kafkaClient(t)
	kc.CreateTopic("test.topic", &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}, false)

	kc.Produce("test.topic", []byte(`{"key":"vaule"}`))

	msg := kc.Consume(ctx, time.Second*3, "test.topic")
	if msg == nil {
		t.Fatalf("consumer message is empty")
	}

	for _, env := range os.Environ() {
		fmt.Printf("-- env: %s\n", env)
	}

	fmt.Println("test success")
}

func TestInitEs(t *testing.T) {
	t.Skip()
	es := steron.ElasticSearch()
	if es == nil {
		t.Fatalf("elasticsearch helper is empty")
	}

	client := es.Client(t)
	if client == nil {
		t.Fatalf("elasticsearch client is empty")
	}
}
