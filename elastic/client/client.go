package client

import (
	"context"
	"fmt"
	"io"
	"testing"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	jsoniter "github.com/json-iterator/go"
)

const indexGeoIp = ".geoip_databases" // system index

type Indices struct {
	Index string `json:"index"`
	//Health       string `json:"health"`
	//Status       string `json:"status"`
	//Uuid         string `json:"uuid"`
	//Pri          string `json:"pri"`
	//Rep          string `json:"rep"`
	//DocsCount    string `json:"docs.count"`
	//DocsDeleted  string `json:"docs.deleted"`
	//StoreSize    string `json:"store.size"`
	//PriStoreSize string `json:"pri.store.size"`
}

type ElasticSearch struct {
	ctx context.Context
	t   *testing.T
	es  *es.Client
}

func NewEs(t *testing.T, address string) *ElasticSearch {
	conf := es.Config{
		Addresses: []string{address},
	}
	esClient, err := es.NewClient(conf)
	if err != nil {
		t.Errorf("new elastic search client error: %s", err)
		return nil
	}
	client := &ElasticSearch{
		ctx: context.Background(),
		t:   t,
		es:  esClient,
	}

	t.Cleanup(func() {
		err = client.cleanup()
		if err != nil {
			t.Errorf("elastic7 cleanup error: %s", err)
		}
	})
	return client
}

func (e *ElasticSearch) Client() *es.Client {
	return e.es
}

// cleanup - delete all indices with him aliases before nex test
func (e *ElasticSearch) cleanup() error {
	res, err := esapi.CatIndicesRequest{Format: "json"}.Do(e.ctx, e.es)
	if err != nil {
		return fmt.Errorf("cat indices error: %s", err)
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body error: %s", err)
	}
	fmt.Println(string(data))

	indices, err := e.indices()
	if err != nil {
		return fmt.Errorf("cleanup aliases error: %s", err)
	}

	if len(indices) == 0 {
		return nil
	}
	res, err = e.es.Indices.Delete(indices)
	if err != nil {
		return fmt.Errorf("delete indices error: %s", err)
	}

	if res.IsError() {
		return fmt.Errorf("delete indices response error: %s", res.String())
	}
	return nil
}

func (e *ElasticSearch) indices() ([]string, error) {
	res, err := esapi.CatIndicesRequest{Format: "json"}.Do(context.Background(), e.es)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	indices := make([]Indices, 0, 10)
	err = jsoniter.Unmarshal(data, &indices)
	if err != nil {
		panic(err)
	}
	result := make([]string, 0, len(indices))
	for _, index := range indices {
		if index.Index == indexGeoIp {
			continue
		}
		result = append(result, index.Index)
	}

	return result, nil
}
