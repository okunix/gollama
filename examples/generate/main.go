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

	req := gollama.GenerateRequest{
		Model:  "gemma3",
		Prompt: "Why is the sky blue?",
	}
	stream, err := client.GenerateStream(ctx, req)
	if err != nil {
		panic(err)
	}

	var metrics gollama.Metrics
	for resp, err := range stream {
		if err != nil {
			panic(err)
		}
		if resp.Done {
			metrics = resp.Metrics
		}
		fmt.Print(resp.Response)
	}
	fmt.Printf("\n%+v\n", metrics)
}
