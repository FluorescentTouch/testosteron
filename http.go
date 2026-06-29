package steron

import (
	"net/http"
	"os"
	"testing"

	"github.com/FluorescentTouch/testosteron/v2/http/client"
	"github.com/FluorescentTouch/testosteron/v2/http/server"
	"github.com/FluorescentTouch/testosteron/v2/sync"
)

type HTTPHelper struct {
	clients sync.Map[WebClient] // t.Name:Client
	servers sync.Map[WebServer] // t.Name:Server

	mainServer WebServer // server started for main init
}

func (h *HTTPHelper) Client(t *testing.T) WebClient {
	if c, ok := h.clients.Get(t.Name()); ok {
		return c
	}

	c := client.NewHTTPClient(t)

	h.clients.Set(t.Name(), c)

	t.Cleanup(func() {
		h.clients.Delete(t.Name())
	})

	return c
}

func (h *HTTPHelper) Server(t *testing.T, envs ...string) WebServer {
	if s, ok := h.servers.Get(t.Name()); ok {
		return s
	}

	s := server.NewHTTPServer(t)

	h.servers.Set(t.Name(), s)

	t.Cleanup(func() {
		h.servers.Delete(t.Name())
	})

	for _, env := range envs {
		_ = os.Setenv(env, s.Addr())
	}

	return s
}

func (h *HTTPHelper) ServerMain(m *testing.M, envs ...string) WebServer {
	h.mainServer = server.NewHTTPMainServer(m)

	for _, env := range envs {
		_ = os.Setenv(env, h.mainServer.Addr())
	}
	return h.mainServer
}

type HandlerCollection struct {
	t *testing.T

	handlers sync.Map[func(http.ResponseWriter, *http.Request)]
}

func New(t *testing.T) *HandlerCollection {
	return &HandlerCollection{
		t:        t,
		handlers: sync.MakeSyncMap[func(http.ResponseWriter, *http.Request)](),
	}
}

func (h *HandlerCollection) Handle(key string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fn, ok := h.handlers.Get(key)
		if !ok {
			h.t.Fatalf("handler not found: %s", key)
		}

		if fn == nil {
			h.t.Fatalf("handler is nil: %s", key)
		}

		fn(w, r)
	}
}

func (h *HandlerCollection) Set(key string, fn func(http.ResponseWriter, *http.Request)) {
	h.handlers.Set(key, fn)
}
