package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"time"
)

type TabsServer struct {
	tabs            *TabStore
	browserMessages chan []byte
	clientId        uint32
	gatewayConn     net.Conn
}

func MakeTabsServer() *TabsServer {
	return &TabsServer{
		tabs:            MakeTabStore(),
		browserMessages: make(chan []byte),
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
		n, err := SendBrowserMsg(s.gatewayConn, listRequest)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Sent %d bytes to gateway", n)
	}()

	return nil
}

func (s *TabsServer) ReadBrowserMsgs() {
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
		s.browserMessages <- buf
	}
}

func (s *TabsServer) HandleBrowserMessages() {
	for buf := range s.browserMessages {
		var rawMsg map[string]json.RawMessage
		err := json.Unmarshal(buf[4:], &rawMsg)
		if err != nil {
			log.Printf("ERROR: unmarshaling msg: %v: `%s`", err, string(buf[4:]))
			continue
		}

		var action string
		if actionRaw, exists := rawMsg["action"]; exists {
			if err := json.Unmarshal(actionRaw, &action); err != nil {
			}
		} else {
			log.Printf("ERROR: Received message with no action")
			continue
		}
		content, exists := rawMsg["content"]
		if !exists {
			log.Printf("ERROR: Received message with no content")
			continue
		}
		switch action {
		case "error":
			var msg string
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			handleBrowserError(msg)

		case "clientId":
			var msg uint32
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Client id is %d", msg)
			s.clientId = msg

		case "created":
			var msg Tab
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Received: %s: %v", action, msg)
			err := s.tabs.Create(&msg)
			if err != nil {
				log.Println(err)
			}

		case "activated":
			var msg ActivatedMsg
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Received: %s: %v", action, msg)
			err := s.tabs.Activate(msg)
			if err != nil {
				log.Println(err)
			}

		case "updated":
			var msg UpdatedMsg
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %#v", err)
				continue
			}
			log.Printf("Received: %s: %d %v", action, msg.TabId, *msg.Delta)
			err := s.tabs.Update(msg)
			if err != nil {
				log.Println(err)
			}

		case "moved":
			var msg MovedMsg
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Received: %s: %v", action, msg)
			err := s.tabs.Move(msg)
			if err != nil {
				log.Println(err)
			}

		case "removed":
			var msg RemovedMsg
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Received: %s: %v", action, msg)
			err := s.tabs.Remove(msg)
			if err != nil {
				log.Println(err)
			}

		case "attached", "detached":
			var msg AttachedMsg
			if err := json.Unmarshal(content, &msg); err != nil {
				log.Printf("ERROR: Failed to parse message content: %v", err)
				continue
			}
			log.Printf("Received: %s: %v", action, msg)
			err := s.tabs.WindowChange(msg)
			if err != nil {
				log.Println(err)
			}

		default:
			log.Printf("ERROR: Msg received from browser with unknown action: %s", action)
		}
	}
}

func handleBrowserError(error string) {
	log.Printf("Browser returned error: %s", error)
}
