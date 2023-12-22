package steron

import (
	"github.com/FluorescentTouch/testosteron/http/client"
	"github.com/FluorescentTouch/testosteron/http/server"
	"github.com/FluorescentTouch/testosteron/sync"
	"testing"
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

func (h *HTTPHelper) Server(t *testing.T) WebServer {
	if s, ok := h.servers.Get(t.Name()); ok {
		return s
	}

	s := server.NewHTTPServer(t)

	h.servers.Set(t.Name(), s)

	t.Cleanup(func() {
		h.servers.Delete(t.Name())
	})

	return s
}

func (h *HTTPHelper) ServerMain(m *testing.M) WebServer {
	h.mainServer = server.NewHTTPMainServer(m)
	return h.mainServer
}
