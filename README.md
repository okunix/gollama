# gollama

A Go client library for the Ollama API.

Manage models, generate completions, chat, and create embeddings against a local or remote Ollama server, with both streaming API support.

## Installation

```bash
go get -u github.com/okunix/gollama
```

Requires a Go version with range-over-func iterator support (Go 1.23+).

## Quick start

```go
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
```

## Client configuration

| Option             | Description                                                                   |
| ---                | ---                                                                           |
| `WithHost(host)`   | Send requests to the given Ollama server host.                                |
| `WithToken(token)` | Authenticate requests with the given token.                                   |
| `WithTimeout(d)`   | Time out requests after duration `d` (applied to the underlying HTTP client). |

```go
client, err := gollama.NewClient(
	ollama.WithHost("https://ollama.example.com"),
	ollama.WithToken(os.Getenv("OLLAMA_TOKEN")),
	ollama.WithTimeout(2*time.Minute),
)
```

## API reference

| Method | Description | Streaming variant |
| --- | --- | --- |
| `NewClient(opts...)` | Create a client; verifies server reachability. | — |
| `Ping()` | Check the server is reachable. | — |
| `Version()` | Get the server version. | — |
| `Tags()` | List local models. | — |
| `Ps()` | List models loaded in memory. | — |
| `ShowModelDetails(model, verbose)` | Modelfile, template, parameters, license. | — |
| `Create(...)` | Create a model. | `CreateStream` |
| `Copy(src, dst)` | Duplicate a model under a new name. | — |
| `Delete(model)` | Remove a model. | — |
| `Pull(model, insecure)` | Download a model. | `PullStream` |
| `Push(model, insecure)` | Upload a model. | `PushStream` |
| `Generate(req)` | Completion for a prompt. | `GenerateStream` |
| `Chat(req)` | Next message in a conversation. | `ChatStream` |
| `Embed(req)` | Generate embeddings. | — |

## Contributing

Issues and pull requests are welcome.

## Related Projects

I'm also building a terraform provider for ollama based on this library <https://github.com/okunix/terraform-provider-ollama>
