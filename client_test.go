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
	stream, err := client.CreateStream(
		t.Context(),
		CreateModel{
			Model:  "alpaca",
			From:   "gemma3",
			System: "You are Alpaca, a helpful AI assistant. You only answer with Emojis.",
		},
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for status, err := range stream {
		if err != nil {
			t.Error(err.Error())
			return
		}
		t.Logf("status: %+v", status.Status)
	}
}

func TestCopy(t *testing.T) {
	ctx := t.Context()
	if err := client.Copy(ctx, "alpaca", "alpaca-copy"); err != nil {
		t.Error(err.Error())
		return
	}
	models, err := client.Tags(ctx)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("Models:")
	for _, v := range models {
		t.Logf("\t%s", v.Name)
	}
	if err := client.Delete(ctx, "alpaca-copy"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestDelete(t *testing.T) {
	err := client.Delete(t.Context(), "alpaca")
	if err != nil {
		t.Error(err.Error())
		return
	}
}
