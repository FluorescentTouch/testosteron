package docker

import (
	"testing"
)

func TestConnRegex(t *testing.T) {
	connection := "postgres://user:passwd@localhost:55001/postgres?"

	_, err := newConnConfig(connection)
	if err != nil {
		t.Errorf(err.Error())
	}
}
