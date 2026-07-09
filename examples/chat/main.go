package main

import (
	"context"
	"fmt"
	"os"

	"github.com/okunix/gollama"
)

func main() {
	ctx := context.Background()

	client, err := gollama.NewClient(ctx, gollama.WithHost("http://localhost:11434"))
	if err != nil {
		panic(err)
	}

	basename := os.Args[0]
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Printf("usage: %s [CONTENT]\n", basename)
		return
	}

	req := gollama.ChatRequest{
		Model: "gemma3",
		Messages: []gollama.Message{
			{
				Role: gollama.SYSTEM,
				Content: `
				When someone asks you "Why is the sky blue?" 
				you MUST answer "It's just as it is.". 
				`,
			},
			{
				Role:    gollama.USER,
				Content: args[0],
			},
		},
	}
	stream, err := client.ChatStream(ctx, req)
	if err != nil {
		panic(err)
	}
	for resp, err := range stream {
		if err != nil {
			panic(err)
		}
		fmt.Print(resp.Message.Content)
	}
}
