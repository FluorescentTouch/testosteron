package steron

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// HTTPServer emulates http Server.
// Do not initialize manualy, use Server(t) instead.
type HTTPServer struct {
	s *httptest.Server
	r *chi.Mux

	t *testing.T
}

func newHTTPServer(t *testing.T) *HTTPServer {
	t.Helper()

	r := chi.NewRouter()
	s := &HTTPServer{s: httptest.NewServer(r), r: r, t: t}
	t.Cleanup(func() {
		s.cleanup()
	})
	return s
}

// HandleFunc adds handler to test server.
func (s *HTTPServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.t.Helper()

	s.r.HandleFunc(pattern, handler)
}

func (s *HTTPServer) Addr() string {
	s.t.Helper()

	return s.s.URL
}

func (s *HTTPServer) cleanup() {
	s.t.Helper()

	s.s.Close()
}

// HTTPMainServer emulates http Server for TestMain.
// Do not initialize manualy, use ServerMain(m) instead.
type HTTPMainServer struct {
	s *httptest.Server
	r *chi.Mux

	m *testing.M // not required, just for difference between HTTPServer
}

func newHTTPMainServer(m *testing.M) *HTTPMainServer {
	r := chi.NewRouter()
	s := &HTTPMainServer{s: httptest.NewServer(r), r: r, m: m}
	return s
}

// HandleFunc adds handler to test server.
func (s *HTTPMainServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.r.HandleFunc(pattern, handler)
}

func (s *HTTPMainServer) Addr() string {
	return s.s.URL
}

func (s *HTTPMainServer) cleanup() {
	s.s.Close()
}
