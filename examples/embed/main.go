package main

import (
	"context"
	"fmt"

	"github.com/okunix/gollama"
)

func main() {
	ctx := context.Background()
	client, err := gollama.NewClient(ctx, gollama.WithHost("http://localhost:11434"))
	if err != nil {
		panic(err)
	}

	model := "embeddinggemma"
	stream, err := client.PullStream(ctx, model, false)
	if err != nil {
		panic(err)
	}

	for status, err := range stream {
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %d/%d\n", status.Status, status.Completed, status.Total)
	}

	req := gollama.EmbedRequest{
		Model: "embeddinggemma",
		Input: "Why is the sky blue? Answer in five words.",
	}
	resp, err := client.Embed(ctx, req)
	if err != nil {
		panic(err)
	}
	fmt.Printf("resp: %+v\n", resp)
}
