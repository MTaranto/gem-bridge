package nativemessaging

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	// MaxMessageBytes limits the size of messages accepted by the native host.
	MaxMessageBytes = 1024 * 1024
)

var (
	// ErrEmptyMessage indicates that a framed message has no JSON payload.
	ErrEmptyMessage = errors.New("native message must not be empty")

	// ErrMessageTooLarge indicates that the payload exceeds the configured limit.
	ErrMessageTooLarge = errors.New("native message exceeds maximum allowed size")
)

// ReadMessage reads one Native Messaging framed payload from reader.
//
// Each message starts with a four-byte unsigned length encoded using the
// platform's native byte order, followed by exactly that number of payload
// bytes.
func ReadMessage(reader io.Reader) ([]byte, error) {
	var header [4]byte

	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, fmt.Errorf("read native message length: %w", err)
	}

	messageSize := binary.NativeEndian.Uint32(header[:])

	if messageSize == 0 {
		return nil, ErrEmptyMessage
	}

	if messageSize > MaxMessageBytes {
		return nil, ErrMessageTooLarge
	}

	payload := make([]byte, messageSize)

	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, fmt.Errorf("read native message payload: %w", err)
	}

	return payload, nil
}

// WriteMessage writes one Native Messaging framed payload to writer.
func WriteMessage(writer io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return ErrEmptyMessage
	}

	if len(payload) > MaxMessageBytes {
		return ErrMessageTooLarge
	}

	var header [4]byte
	binary.NativeEndian.PutUint32(header[:], uint32(len(payload)))

	if err := writeAll(writer, header[:]); err != nil {
		return fmt.Errorf("write native message length: %w", err)
	}

	if err := writeAll(writer, payload); err != nil {
		return fmt.Errorf("write native message payload: %w", err)
	}

	return nil
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}

		if written == 0 {
			return io.ErrShortWrite
		}

		data = data[written:]
	}

	return nil
}
