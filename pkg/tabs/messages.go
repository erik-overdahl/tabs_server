package tabs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
)

type Message struct {
	ID      uuid.UUID       `json:"id"`
	Action  string          `json:"action"`
	Content json.RawMessage `json:"content"`
}

type EmptyObj map[string]any

func MakeMessage(action string, content any) *Message {
	contentBuf, _ := json.Marshal(content)
	return &Message{
		ID:      uuid.New(),
		Action:  action,
		Content: contentBuf,
	}
}

func ReadMsg(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg size: %v", err)
	}
	msgSize := int(binary.LittleEndian.Uint32(buf))
	buf = append(buf, make([]byte, msgSize)...)
	_, err = io.ReadFull(r, buf[4:])
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg: %v", err)
	}
	msg := &Message{}
	if err := json.Unmarshal(buf[4:], msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func SendMsg(w io.Writer, msg Message) error {
	// add size if not already there
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(buf)))
	buf = append(size, buf...)
	_, err = w.Write(buf)
	return err
}

func CastAndCall[T any](content json.RawMessage, f func(*T) error) error {
	var msg T
	if err := json.Unmarshal(content, &msg); err != nil {
		return fmt.Errorf("ERROR: Failed to parse message content: %v", err)
	}
	log.Printf("RECEIVED: %#v", msg)
	return f(&msg)
}

func handleBrowserError(err *string) error {
	log.Printf("Browser returned error: %s", *err)
	return nil
}
