package gateway

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	// connection to the Gateway
	conn net.Conn
	// routing for responses
	requests map[uuid.UUID]chan *Response
}

func (this *Client) Connect(sockAddr string) error {
	if conn, err := net.Dial("unix", sockAddr); err != nil {
		return err
	} else {
		this.conn = conn
	}
	go this.listen()
	return nil
}

func (this *Client) Request(msg *Request) (*Response, error) {
	if this.conn == nil {
		return nil, errors.New("cannot send request to closed gateway connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	responseChan := make(chan *Response)
	msg.Id = uuid.New()
	this.requests[msg.Id] = responseChan

	SendMsg(this.conn, &Message{Request: msg})

	select {
	case <-ctx.Done():
		return nil, errors.New("request timed out after 5 seconds")
	case response := <-responseChan:
		delete(this.requests, msg.Id)
		return response, nil
	}
}

func (this *Client) Subscribe(eventName string) (*Response, error) {
	return this.Request(&Request{Method: "subscribe", Data: []byte(eventName)})
}

func (this *Client) listen() {
	for {
		msg, err := ReadMsg(this.conn)
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			log.Printf("ERROR: %#v", err)
			continue
		}

		switch {
		case msg.Response != nil:
			response := msg.Response
			log.Printf("Received response for %s", response.Id)
			if responseChan, exists := this.requests[response.Id]; exists {
				responseChan <- response
			} else {
				log.Printf("Received unexpected msg response: %v", response)
			}
		case msg.Event != nil:

		default:
			log.Printf("Received unexpected msg: %v", msg)
		}
	}
}
