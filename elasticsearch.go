package steron

import (
	"fmt"
	"testing"

	"github.com/FluorescentTouch/testosteron/v2/elastic/client"
	"github.com/FluorescentTouch/testosteron/v2/elastic/docker"
	"github.com/FluorescentTouch/testosteron/v2/sync"
	tc "github.com/testcontainers/testcontainers-go"
)

type ElasticSearchService struct {
	opts []tc.ContainerCustomizer
}

func NewElasticSearchService(opts ...tc.ContainerCustomizer) *ElasticSearchService {
	return &ElasticSearchService{opts: opts}
}

func (e *ElasticSearchService) WithHelper(h *Helper) error {
	esContainer, err := docker.RunContainer(e.opts...)
	if err != nil {
		return fmt.Errorf("could not start elasticsearch container: %w", err)
	}
	h.elasticSearch.container = esContainer
	h.elasticSearch.opts = e.opts
	h.cfg.esHost = esContainer.Host()
	return nil
}

type ElasticSearchHelper struct {
	clients   sync.Map[ElasticSearchClient]
	opts      []tc.ContainerCustomizer
	container *docker.ElasticsearchContainer
}

func (h *ElasticSearchHelper) Client(t *testing.T) ElasticSearchClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	es := h.container

	if es == nil {
		e, err := docker.RunContainer(h.opts...)
		if err != nil {
			t.Errorf("start container ES err: %s", err)
			return nil
		}

		t.Cleanup(func() {
			err = e.Cleanup()
			if err != nil {
				t.Errorf("es cleanup err: %s", err)
			}
		})

		es = e
	}

	c := client.NewEs(t, es.Host())
	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}
