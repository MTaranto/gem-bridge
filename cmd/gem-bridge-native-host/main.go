package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mtaranto/gem-bridge/internal/nativemessaging"
	"github.com/mtaranto/gem-bridge/internal/security"
	"github.com/mtaranto/gem-bridge/internal/tools"
)

const workspaceEnvName = "GEM_BRIDGE_WORKSPACE"

// Request represents a message received from the browser extension.
type Request struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
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

	case "readFile":
		return readFile(request)

	default:
		return Response{
			Success: false,
			Error:   "unsupported request type: " + request.Type,
		}
	}
}

func readFile(request Request) Response {
	if request.Path == "" {
		return Response{
			Success: false,
			Error:   "path is required for readFile",
		}
	}

	workspaceRoot := os.Getenv(workspaceEnvName)
	if workspaceRoot == "" {
		return Response{
			Success: false,
			Error:   workspaceEnvName + " is not configured",
		}
	}

	workspace, err := security.NewWorkspace(workspaceRoot)
	if err != nil {
		return Response{
			Success: false,
			Error:   "configure workspace: " + err.Error(),
		}
	}

	fileTools := tools.NewFileTools(workspace)

	content, err := fileTools.ReadFile(request.Path)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Type:    "fileContent",
		Data: map[string]interface{}{
			"path":    request.Path,
			"content": content,
			"size":    len([]byte(content)),
		},
	}
}
