package tabs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
)

type Response struct {
	ID     uuid.UUID       `json:"id"`
	Status string          `json:"status"`
	Info   json.RawMessage `json:"info,omitempty"`
}

type Request struct {
	ID     uuid.UUID `json:"id"`
	Method string    `json:"method"`
	TabId  int       `json:"tabId,omitempty"`
	TabIds []int     `json:"tabIds,omitempty"`
	Props  any       `json:"props,omitempty"`
}

type rawMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Message struct {
	Response *Response
	Request  *Request
	Event    Event
}

func (msg *Message) MarshalJSON() ([]byte, error) {
	var msgType string
	var content any
	switch {
	case msg.Request != nil:
		msgType = "request"
		content = msg.Request
	case msg.Response != nil:
		msgType = "response"
		content = msg.Response
	case msg.Event != nil:
		msgType = "event"
		log.Printf("Marshaling event: %v", msg.Event)
		eventBytes, err := json.Marshal(msg.Event)
		if err != nil {
			return nil, err
		}
		content = rawMessage{Type: msg.Event.Name(), Data: eventBytes}
	}
	data, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rawMessage{Type: msgType, Data: data})
}

func (msg *Message) UnmarshalJSON(data []byte) error {
	var raw rawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch raw.Type {
	case "request":
		msg.Request = &Request{}
		return json.Unmarshal(raw.Data, msg.Request)
	case "response":
		msg.Response = &Response{}
		return json.Unmarshal(raw.Data, msg.Response)
	case "event":
		var rawEvent rawMessage
		if err := json.Unmarshal(raw.Data, &rawEvent); err != nil {
			return err
		}
		var event Event
		switch rawEvent.Type {
		case "activated":
			event = &ActivatedMsg{}
		case "updated":
			event = &UpdatedMsg{}
		case "created":
			event = &CreatedMsg{}
		case "removed":
			event = &RemovedMsg{}
		case "moved":
			event = &MovedMsg{}
		case "attached", "detached":
			event = &AttachedMsg{}
		default:
			return fmt.Errorf("Event of unknown type: %s", rawEvent.Type)
		}
		log.Printf("Unmarshaling to event: %s", string(raw.Data))
		if err := json.Unmarshal(rawEvent.Data, &event); err != nil {
			return err
		}
		log.Printf("Unmarshalled to: %v", event)
		msg.Event = event
		return nil
	default:
		return fmt.Errorf("Message of unknown type: %s", raw.Type)
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
		log.Println(err)
		return nil, err
	}
	log.Println("RECEIVED:", msg)
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
