package srv

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	steron "github.com/FluorescentTouch/testosteron/v2"
)

func TestMain(m *testing.M) {
	// init package with required options
	_, err := steron.Init()
	if err != nil {
		panic(err)
	}

	// run the app
	code := m.Run()

	steron.Cleanup()
	os.Exit(code)
}

func TestSrv(t *testing.T) {
	srv := steron.HTTP().Server(t)

	srv.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/me +-111")
	})

	steron.HTTP().Client(t).Get(srv.Addr() + "/me")

	srv.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/me +-222")
	})

	steron.HTTP().Client(t).Get(srv.Addr() + "/me")

	srv.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/me +-333")
	})

	steron.HTTP().Client(t).Get(srv.Addr() + "/me")
}
