package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type TabsServer struct {
	tabs            *TabStore
	clientId        uint32
	gatewayConn     net.Conn
}

func MakeTabsServer() *TabsServer {
	return &TabsServer{
		tabs:            MakeTabStore(),
		clientId:        0,
	}
}

func (s *TabsServer) ConnectBrowserGateway() error {
	c, err := net.Dial("unix", GatewaySockAddr)
	if err != nil {
		return err
	}
	s.gatewayConn = c
	log.Printf("Connected to browser gateway")
	go func() {
		log.Println("Waiting for client id...")
		for s.clientId == 0 {
			time.Sleep(time.Second)
		}

		var query map[string]any = map[string]any{
			"action":   "query",
			"clientId": int(s.clientId),
			"content":  map[string]any{},
		}
		listRequest, err := json.Marshal(query)
		if err != nil {
			log.Printf("ERROR: Unable to build list request: %v", err)
		}
		_, err = SendBrowserMsg(s.gatewayConn, listRequest)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func (s *TabsServer) ReadBrowserMsgs() {
	messages := make(chan []byte)

	go func(msgs chan []byte) {
		for m := range msgs {
			err := s.HandleBrowserMessage(m)
			if err != nil {
				log.Println(err)
			}
		}
	}(messages)

	for {
		buf, err := ReadBrowserMsg(s.gatewayConn)
		if err == io.EOF {
			log.Printf("Received EOF: Gateway hangup")
			s.gatewayConn = nil
			return
		} else if err != nil {
			log.Printf("ERROR: reading msg: %#v", err)
			continue
		}
		messages <- buf
	}
}

func (s *TabsServer) HandleBrowserMessage(buf []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(buf[4:], &rawMsg); err != nil {
		log.Printf("ERROR: unmarshaling msg: %v: `%s`", err, string(buf[4:]))
		return nil
	}

	var action string
	if actionRaw, exists := rawMsg["action"]; !exists {
		return fmt.Errorf("ERROR: Received message with no action")
	} else if err := json.Unmarshal(actionRaw, &action); err != nil {
		return fmt.Errorf("ERROR: unmarshaling msg action: %v", err)
	}
	content, exists := rawMsg["content"]
	if !exists {
		log.Printf("ERROR: Received message with no content")
		return nil
	}

	switch action {
	case "error":
		return CastAndCall(content, handleBrowserError)

	case "clientId":
		return CastAndCall(content,
			func(n *uint32) error {
				s.clientId = *n
				return nil
			})

	case "created":
		return CastAndCall(content, s.tabs.Create)

	case "activated":
		return CastAndCall(content, s.tabs.Activate)

	case "updated":
		return CastAndCall(content, s.tabs.Update)

	case "moved":
		return CastAndCall(content, s.tabs.Move)

	case "removed":
		return CastAndCall(content, s.tabs.Remove)

	case "attached", "detached":
		return CastAndCall(content, s.tabs.WindowChange)

	default:
		return fmt.Errorf("ERROR: Msg received from browser with unknown action: %s", action)
	}
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
