package gollama

import (
	"context"
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

func TestTags(t *testing.T) {
	models, err := client.Tags(t.Context())
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("models: %+v", models)
}

func TestPS(t *testing.T) {
	models, err := client.Ps(t.Context())
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("ps: %+v", models)
}

func TestDetails(t *testing.T) {
	models, err := client.ShowModelDetails(t.Context(), "gemma3", false)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("details: %+v", models)
}

func TestCreate(t *testing.T) {
	model := CreateModel{
		Model:  "alpaca",
		From:   "gemma3",
		System: "You are Alpaca, a helpful AI assistant. You only answer with Emojis.",
	}
	err := client.Create(t.Context(), model)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestCreateStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	streamChan, errChan := client.CreateStream(
		ctx,
		CreateModel{
			Model:  "alpaca",
			From:   "gemma3",
			System: "You are Alpaca, a helpful AI assistant. You only answer with Emojis.",
		},
	)

Outer:
	for {
		select {
		case status, ok := <-streamChan:
			if !ok {
				break Outer
			}
			t.Logf("status log: %+v", status)
		case err := <-errChan:
			t.Error(err.Error())
			break Outer
		}
	}
}

func TestDelete(t *testing.T) {
	err := client.Delete(t.Context(), "alpaca")
	if err != nil {
		t.Error(err.Error())
		return
	}
}
