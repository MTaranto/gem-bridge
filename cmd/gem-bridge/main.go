package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mtaranto/gem-bridge/internal/security"
	"github.com/mtaranto/gem-bridge/internal/tools"
)

// Request represents a single tool invocation received by the daemon.
//
// In this first version, requests are passed through the command line as JSON.
// Later, this same structure can be reused by a WebSocket server, an HTTP API,
// or a browser Native Messaging bridge.
type Request struct {
	Tool    string  `json:"tool"`
	Path    string  `json:"path"`
	Content *string `json:"content,omitempty"`
}

// Response represents the standard JSON response returned by every tool.
//
// Keeping a consistent response format makes it easier for browser extensions,
// local scripts, and future AI integrations to consume daemon results.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	var workspaceRoot string

	flag.StringVar(
		&workspaceRoot,
		"workspace",
		".",
		"workspace root directory used as the filesystem security boundary",
	)

	flag.Parse()

	if flag.NArg() < 1 {
		printJSON(Response{
			Success: false,
			Error:   "usage: gem-bridge [--workspace path] '<json-request>'",
		})
		os.Exit(1)
	}

	workspace, err := security.NewWorkspace(workspaceRoot)
	if err != nil {
		printJSON(Response{
			Success: false,
			Error:   err.Error(),
		})
		os.Exit(1)
	}

	fileTools := tools.NewFileTools(workspace)
	gitTools := tools.NewGitTools(workspace)

	var req Request
	if err := json.Unmarshal([]byte(flag.Arg(0)), &req); err != nil {
		printJSON(Response{
			Success: false,
			Error:   "invalid JSON: " + err.Error(),
		})
		os.Exit(1)
	}

	switch req.Tool {
	case "listDirectory":
		data, err := fileTools.ListDirectory(req.Path)
		if err != nil {
			printJSON(Response{
				Success: false,
				Error:   err.Error(),
			})
			os.Exit(1)
		}

		printJSON(Response{
			Success: true,
			Data:    data,
		})

	case "readFile":
		data, err := fileTools.ReadFile(req.Path)
		if err != nil {
			printJSON(Response{
				Success: false,
				Error:   err.Error(),
			})
			os.Exit(1)
		}

		printJSON(Response{
			Success: true,
			Data:    data,
		})

	case "writeFile":
		if req.Content == nil {
			printJSON(Response{
				Success: false,
				Error:   "content is required for writeFile",
			})
			os.Exit(1)
		}

		if err := fileTools.WriteFile(req.Path, *req.Content); err != nil {
			printJSON(Response{
				Success: false,
				Error:   err.Error(),
			})
			os.Exit(1)
		}

		printJSON(Response{
			Success: true,
			Data: map[string]string{
				"path": req.Path,
			},
		})

	case "gitStatus":
		data, err := gitTools.StatusShort()
		if err != nil {
			printJSON(Response{
				Success: false,
				Error:   err.Error(),
			})
			os.Exit(1)
		}

		printJSON(Response{
			Success: true,
			Data:    data,
		})

	default:
		printJSON(Response{
			Success: false,
			Error:   "unknown tool: " + req.Tool,
		})
		os.Exit(1)
	}
}

// printJSON writes a formatted JSON response to standard output.
func printJSON(resp Response) {
	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Println(`{"success":false,"error":"failed to encode JSON response"}`)
		return
	}

	fmt.Println(string(output))
}
