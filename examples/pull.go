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

	model := "gemma3:270m"
	if err := client.Delete(ctx, model); err != nil {
		panic(err)
	}

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
}
