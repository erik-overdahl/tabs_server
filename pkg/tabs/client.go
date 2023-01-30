package tabs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
)

func ptr[T any](b T) *T { return &b }

// The property structs expected as args by the browser
// Pointers are used to denote optional fields
// Use the `ptr` helper function to create pointers to primitives
// e.g. &CreateProperties{Active: ptr(true)}

type CreateProperties struct {
	WindowId         *int    `json:"windowId,omitempty"`
	Index            *int    `json:"index,omitempty"`
	Url              *string `json:"url,omitempty"`
	Active           *bool   `json:"active,omitempty"`
	Pinned           *bool   `json:"pinned,omitempty"`
	OpenerTabId      *int    `json:"openerTabId,omitempty"`
	CookieStoreId    *int    `json:"cookieStoreId,omitempty"`
	OpenInReaderMode *bool   `json:"openInReaderMode,omitempty"`
	Discarded        *bool   `json:"discarded,omitempty"`
	Title            *string `json:"title,omitempty"`
	Muted            *bool   `json:"muted,omitempty"`
}

type UpdateProperties struct {
	Url         *string `json:"url,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	Highlighted *bool   `json:"highlighted,omitempty"`
	Pinned      *bool   `json:"pinned,omitempty"`
	Muted       *bool   `json:"muted,omitempty"`
	OpenerTabId *int    `json:"openerTabId,omitempty"`
	LoadReplace *bool   `json:"loadReplace,omitempty"`
	SuccessorId *int    `json:"successorId,omitempty"`
}

type DuplicateProperties struct {
	Index  *int  `json:"index,omitempty"`
	Active *bool `json:"active,omitempty"`
}

type MoveProperties struct {
	WindowId int `json:"windowId,omitempty"`
	Index    int `json:"index,omitempty"`
}

type ReloadProperties struct {
	BypassCache bool `json:"bypassCache,omitempty"`
}

type TabsClient struct {
	Updates     chan *Message
	gatewayConn net.Conn
	requests    map[uuid.UUID]chan *Message
}

func MakeTabsClient() *TabsClient {
	return &TabsClient{Updates: make(chan *Message), requests: map[uuid.UUID]chan *Message{}}
}

func (client *TabsClient) ConnectBrowserGateway() error {
	conn, err := net.Dial("unix", GatewaySockAddr)
	if err != nil {
		return err
	}
	client.gatewayConn = conn
	go client.listen()
	log.Printf("Connected to browser gateway")
	return nil
}

func (client *TabsClient) Activate(tabId int) error {
	_, err := client.Request(*MakeMessage(
		"update",
		map[string]any{
			"tabId": tabId,
			"delta": map[string]any{
				"active": true,
			},
		},
	))
	return err
}

func (client *TabsClient) Create(props CreateProperties) (int, error) {
	resp, err := client.Request(*MakeMessage("create", props))
	if err != nil {
		return -1, err
	}
	var tabId int
	if err := json.Unmarshal(resp.Content, &tabId); err != nil {
		return -1, err
	}
	return tabId, nil
}

func (client *TabsClient) Duplicate(tabId int, props *DuplicateProperties) (int, error) {
	resp, err := client.Request(
		*MakeMessage(
			"duplicate",
			map[string]any{
				"tabId": tabId,
				"props": props,
			}))
	if err != nil {
		return -1, err
	}
	var newId int
	if err := json.Unmarshal(resp.Content, &newId); err != nil {
		return -1, err
	}
	return newId, nil
}

func (client *TabsClient) Update(tabId int, props *UpdateProperties) error {
	_, err := client.Request(
		*MakeMessage(
			"update",
			map[string]any{
				"tabId": tabId,
				"props": props,
			}))
	return err
}

func (client *TabsClient) Move(tabId int, props MoveProperties) error {
	_, err := client.Request(
		*MakeMessage(
			"move",
			map[string]any{
				"tabId": tabId,
				"props": props,
			}))
	return err
}

func (client *TabsClient) Reload(tabId int, props *ReloadProperties) error {
	_, err := client.Request(
		*MakeMessage(
			"reload",
			map[string]any{
				"tabId": tabId,
				"props": props,
			}))
	return err
}

func (client *TabsClient) Close(tabId int, tabIds ...int) error {
	_, err := client.Request(*MakeMessage("remove", append(tabIds, tabId)))
	return err
}

func (client *TabsClient) Discard(tabId int, tabIds ...int) error {
	_, err := client.Request(*MakeMessage("discard", append(tabIds, tabId)))
	return err
}


func (client *TabsClient) Hide(tabId int, tabIds ...int) error {
	_, err := client.Request(*MakeMessage("hide", append(tabIds, tabId)))
	return err
}

func (client *TabsClient) Show(tabId int, tabIds ...int) error {
	_, err := client.Request(*MakeMessage("show", append(tabIds, tabId)))
	return err
}

func (client *TabsClient) ToggleReaderMode(tabId int) error {
	_, err := client.Request(*MakeMessage("toggleReaderMode", tabId))
	return err
}

func (client *TabsClient) GoBack(tabId int) error {
	_, err := client.Request(*MakeMessage("goBack", tabId))
	return err
}

func (client *TabsClient) GoForward(tabId int) error {
	_, err := client.Request(*MakeMessage("goForward", tabId))
	return err
}

func (client *TabsClient) Request(msg Message) (*Message, error) {
	if client.gatewayConn == nil {
		return nil, errors.New("Cannot send request to closed gateway connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	responseChan := make(chan *Message)
	client.requests[msg.ID] = responseChan
	SendMsg(client.gatewayConn, msg)

	select {
	case <-ctx.Done():
		return nil, errors.New("Request timed out after 5 seconds")
	case response := <-responseChan:
		delete(client.requests, msg.ID)
		return response, nil
	}
}

func (client *TabsClient) listen() {
	for {
		msg, err := ReadMsg(client.gatewayConn)
		if err == io.EOF {
			log.Fatalf("Received EOF from gateway")
		} else if err != nil {
			log.Printf("ERROR: %#v", err)
			continue
		}
		if msg.ID != uuid.Nil {
			log.Printf("Received response for %s", msg.ID)
			client.requests[msg.ID] <- msg
		} else {
			client.Updates <- msg
		}
	}
}
