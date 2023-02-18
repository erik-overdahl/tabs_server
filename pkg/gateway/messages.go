package gateway

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	STATUS_SUCCESS = RequestStatus("success")
	STATUS_FAILURE = RequestStatus("fail")
)

type Request struct {
	Id     uuid.UUID `json:"id"`
	Method string    `json:"method"`
	Data   any       `json:"data"`
}

type Response struct {
	Id     uuid.UUID       `json:"id"`
	Status RequestStatus   `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Message struct {
	Response *Response `json:"response,omitempty"`
	Request  *Request  `json:"request,omitempty"`
	Event    *Event    `json:"event,omitempty"`
}

func SendMsg(conn io.Writer, msg *Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	wire := make([]byte, len(buf)+4, len(buf)+4)
	binary.LittleEndian.PutUint32(wire, uint32(len(buf)))
	for i := range buf {
		wire[i+4] = buf[i]
	}
	_, err = conn.Write(wire)
	return err
}

func ReadMsg(conn io.Reader) (*Message, error) {
	var msgSize int
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg size: %v", err)
	} else {
		msgSize = int(binary.LittleEndian.Uint32(buf))
	}

	msg := &Message{}
	buf = append(buf, make([]byte, msgSize)...)
	if _, err := io.ReadFull(conn, buf[4:]); err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("ERROR: reading msg: %v", err)
	} else if err := json.Unmarshal(buf[4:], msg); err != nil {
		return nil, err
	}
	return msg, nil
}
