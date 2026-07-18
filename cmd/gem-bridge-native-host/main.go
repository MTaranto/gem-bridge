package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mtaranto/gem-bridge/internal/nativemessaging"
)

// Request represents a message received from the browser extension.
type Request struct {
	Type string `json:"type"`
}

// Response represents a message returned to the browser extension.
type Response struct {
	Success bool        `json:"success"`
	Type    string      `json:"type,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gem-bridge-native-host: %v\n", err)
		os.Exit(1)
	}
}

func run(reader io.Reader, writer io.Writer) error {
	payload, err := nativemessaging.ReadMessage(reader)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}

	response := handleRequest(payload)

	encodedResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}

	if err := nativemessaging.WriteMessage(writer, encodedResponse); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func handleRequest(payload []byte) Response {
	var request Request

	if err := json.Unmarshal(payload, &request); err != nil {
		return Response{
			Success: false,
			Error:   "invalid JSON request",
		}
	}

	if request.Type == "" {
		return Response{
			Success: false,
			Error:   "request type is required",
		}
	}

	switch request.Type {
	case "ping":
		return Response{
			Success: true,
			Type:    "pong",
			Data: map[string]string{
				"host": "gem-bridge-native-host",
			},
		}

	default:
		return Response{
			Success: false,
			Error:   "unsupported request type: " + request.Type,
		}
	}
}
