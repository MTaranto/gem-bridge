package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/mtaranto/gem-bridge/internal/nativemessaging"
)

func TestRunRespondsToPing(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":"ping"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected native host to succeed: %v", err)
	}

	response := readResponse(t, &output)

	if !response.Success {
		t.Fatalf("expected successful response, got error %q", response.Error)
	}

	if response.Type != "pong" {
		t.Fatalf("expected response type pong, got %q", response.Type)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %#v", response.Data)
	}

	if data["host"] != "gem-bridge-native-host" {
		t.Fatalf(
			"expected host gem-bridge-native-host, got %#v",
			data["host"],
		)
	}
}

func TestRunReturnsErrorResponseForInvalidJSON(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected malformed JSON to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "invalid JSON request" {
		t.Fatalf("expected invalid JSON error, got %q", response.Error)
	}
}

func TestRunReturnsErrorResponseWhenTypeIsMissing(t *testing.T) {
	input := frameMessage(t, []byte(`{}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected missing type to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "request type is required" {
		t.Fatalf("expected missing type error, got %q", response.Error)
	}
}

func TestRunReturnsErrorResponseForUnsupportedType(t *testing.T) {
	input := frameMessage(t, []byte(`{"type":"unknown"}`))

	var output bytes.Buffer

	if err := run(&input, &output); err != nil {
		t.Fatalf("expected unsupported type to produce a response: %v", err)
	}

	response := readResponse(t, &output)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	expectedError := "unsupported request type: unknown"
	if response.Error != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, response.Error)
	}
}

func TestRunFailsForIncompleteFrame(t *testing.T) {
	input := bytes.NewBuffer([]byte{1, 2})

	var output bytes.Buffer

	err := run(input, &output)
	if err == nil {
		t.Fatal("expected incomplete frame to fail")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}

	if !strings.Contains(err.Error(), "read request:") {
		t.Fatalf("expected read request context, got %q", err.Error())
	}
}

func frameMessage(t *testing.T, payload []byte) bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer

	if err := nativemessaging.WriteMessage(&buffer, payload); err != nil {
		t.Fatalf("expected message framing to succeed: %v", err)
	}

	return buffer
}

func readResponse(t *testing.T, reader io.Reader) Response {
	t.Helper()

	payload, err := nativemessaging.ReadMessage(reader)
	if err != nil {
		t.Fatalf("expected framed response read to succeed: %v", err)
	}

	var response Response

	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("expected response JSON decoding to succeed: %v", err)
	}

	return response
}
