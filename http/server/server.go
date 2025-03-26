package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func NewHTTPServer(t *testing.T) *HTTPServer {
	t.Helper()

	r := chi.NewRouter()
	s := &HTTPServer{s: httptest.NewServer(r), r: r, t: t}
	t.Cleanup(func() {
		s.Cleanup()
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

func (s *HTTPServer) Cleanup() {
	s.t.Helper()

	s.s.Close()
}

func (s *HTTPServer) NewDebugHandler(f http.HandlerFunc) func(comments ...string) {
	mx := make(chan struct{})

	s.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		f(w, r)

		select {
		case mx <- struct{}{}:
		default:
			return
		}
	})

	return func(comments ...string) {
		s.t.Logf("----- Go on %s/next to next breakpoint, %s\n", s.Addr(), strings.Join(comments, ","))
		<-mx
	}
}

// HTTPMainServer emulates http Server for TestMain.
// Do not initialize manualy, use ServerMain(m) instead.
type HTTPMainServer struct {
	s *httptest.Server
	r *chi.Mux

	m *testing.M // not required, just for difference between HTTPServer
}

func NewHTTPMainServer(m *testing.M) *HTTPMainServer {
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

func (s *HTTPMainServer) Cleanup() {
	s.s.Close()
}

func (s *HTTPMainServer) NewDebugHandler(f http.HandlerFunc) func(comments ...string) {
	mx := make(chan struct{})

	s.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		f(w, r)

		select {
		case mx <- struct{}{}:
		default:
			return
		}
	})

	return func(comments ...string) {
		println("----- Go on", s.Addr(), "/next to next breakpoint", strings.Join(comments, ","))
		<-mx
	}
}
