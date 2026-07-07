package gollama

import (
	"testing"
	"time"
)

var client *Client

func TestNewClient(t *testing.T) {
	var err error
	client, err = NewClient(
		t.Context(),
		WithHost("http://localhost:11434"),
		WithTimeout(time.Second),
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestPing(t *testing.T) {
	if err := client.Ping(t.Context()); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestVersion(t *testing.T) {
	version, err := client.Version(t.Context())
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("ollama version: " + version)
}
