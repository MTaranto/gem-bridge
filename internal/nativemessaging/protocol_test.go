package nativemessaging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestReadMessageReturnsPayload(t *testing.T) {
	payload := []byte(`{"type":"ping"}`)

	var input bytes.Buffer
	writeHeader(t, &input, uint32(len(payload)))

	if _, err := input.Write(payload); err != nil {
		t.Fatalf("expected payload write to succeed: %v", err)
	}

	message, err := ReadMessage(&input)
	if err != nil {
		t.Fatalf("expected message read to succeed: %v", err)
	}

	if !bytes.Equal(message, payload) {
		t.Fatalf("expected payload %q, got %q", payload, message)
	}
}

func TestReadMessageRejectsEmptyPayload(t *testing.T) {
	var input bytes.Buffer
	writeHeader(t, &input, 0)

	_, err := ReadMessage(&input)
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage, got %v", err)
	}
}

func TestReadMessageRejectsOversizedPayload(t *testing.T) {
	var input bytes.Buffer
	writeHeader(t, &input, MaxMessageBytes+1)

	_, err := ReadMessage(&input)
	if !errors.Is(err, ErrMessageTooLarge) {
		t.Fatalf("expected ErrMessageTooLarge, got %v", err)
	}
}

func TestReadMessageFailsForIncompleteHeader(t *testing.T) {
	input := bytes.NewBuffer([]byte{1, 2})

	_, err := ReadMessage(input)
	if err == nil {
		t.Fatal("expected incomplete header to fail")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestReadMessageFailsForIncompletePayload(t *testing.T) {
	var input bytes.Buffer
	writeHeader(t, &input, 5)

	if _, err := input.Write([]byte("abc")); err != nil {
		t.Fatalf("expected partial payload write to succeed: %v", err)
	}

	_, err := ReadMessage(&input)
	if err == nil {
		t.Fatal("expected incomplete payload to fail")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestWriteMessageWritesFramedPayload(t *testing.T) {
	payload := []byte(`{"type":"pong"}`)

	var output bytes.Buffer

	if err := WriteMessage(&output, payload); err != nil {
		t.Fatalf("expected message write to succeed: %v", err)
	}

	frame := output.Bytes()

	if len(frame) != 4+len(payload) {
		t.Fatalf(
			"expected frame size %d, got %d",
			4+len(payload),
			len(frame),
		)
	}

	messageSize := binary.NativeEndian.Uint32(frame[:4])
	if messageSize != uint32(len(payload)) {
		t.Fatalf(
			"expected encoded payload size %d, got %d",
			len(payload),
			messageSize,
		)
	}

	if !bytes.Equal(frame[4:], payload) {
		t.Fatalf("expected payload %q, got %q", payload, frame[4:])
	}
}

func TestWriteMessageRejectsEmptyPayload(t *testing.T) {
	var output bytes.Buffer

	err := WriteMessage(&output, nil)
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage, got %v", err)
	}
}

func TestWriteMessageRejectsOversizedPayload(t *testing.T) {
	var output bytes.Buffer
	payload := make([]byte, MaxMessageBytes+1)

	err := WriteMessage(&output, payload)
	if !errors.Is(err, ErrMessageTooLarge) {
		t.Fatalf("expected ErrMessageTooLarge, got %v", err)
	}
}

func TestWriteMessageFailsWhenWriterFails(t *testing.T) {
	expectedErr := errors.New("write failed")
	writer := &failingWriter{
		err: expectedErr,
	}

	err := WriteMessage(writer, []byte(`{"type":"pong"}`))
	if err == nil {
		t.Fatal("expected writer failure")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped writer error, got %v", err)
	}
}

func writeHeader(t *testing.T, writer io.Writer, messageSize uint32) {
	t.Helper()

	var header [4]byte
	binary.NativeEndian.PutUint32(header[:], messageSize)

	if _, err := writer.Write(header[:]); err != nil {
		t.Fatalf("expected header write to succeed: %v", err)
	}
}

type failingWriter struct {
	err error
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}
