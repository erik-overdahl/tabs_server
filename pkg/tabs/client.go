package tabs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Updates     chan Event
	gatewayConn net.Conn
	requests    map[uuid.UUID]chan *Response
}

func MakeTabsClient() *TabsClient {
	return &TabsClient{Updates: make(chan Event), requests: map[uuid.UUID]chan *Response{}}
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

func (client *TabsClient) GetList() ([]*Tab, error) {
	response, err := client.Request(&Request{
		Method: "list",
	})
	if err != nil {
		return nil, err
	} else if response.Status != "success" {
		return nil, fmt.Errorf("Browser responded: %s: %s", response.Status, string(response.Info))
	}
	var tabList []*Tab
	if err := json.Unmarshal(response.Info, &tabList); err != nil {
		return nil, err
	}
	return tabList, nil
}

func (client *TabsClient) Activate(tabId int) error {
	if response, err := client.Request(&Request{
		Method: "update",
		TabId: tabId,
		Props: &UpdateProperties{Active: ptr(true)},
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %s", response.Status, string(response.Info))
	}
	return nil
}

func (client *TabsClient) Create(props CreateProperties) (int, error) {
	response, err := client.Request(&Request{
		Method: "create",
		Props: &props,
	})
	if err != nil {
		return -1, err
	} else if response.Status != "success" {
		return -1, fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	var newTabId int
	if err := json.Unmarshal(response.Info, &newTabId); err != nil {
		return -1, err
	}
	return newTabId, nil
}

func (client *TabsClient) Duplicate(tabId int, props *DuplicateProperties) (int, error) {
	response, err := client.Request(&Request{
		Method: "duplicate",
		TabId: tabId,
		Props: props,
	})
	if err != nil {
		return -1, err
	} else if response.Status != "success" {
		return -1, fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	var newTabId int
	if err := json.Unmarshal(response.Info, &newTabId); err != nil {
		return -1, err
	}
	return newTabId, nil
}

func (client *TabsClient) Update(tabId int, props *UpdateProperties) error {
	if response, err := client.Request(&Request{
		Method: "update",
		Props: props,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Move(tabId int, props MoveProperties) error {
	if response, err := client.Request(&Request{
		Method: "move",
		TabId: tabId,
		Props: &props,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Reload(tabId int, props *ReloadProperties) error {
	if response, err := client.Request(&Request{
		Method: "reload",
		TabId: tabId,
		Props: &props,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Close(tabId int, tabIds ...int) error {
	if response, err := client.Request(&Request{
		Method: "remove",
		TabIds: append(tabIds, tabId),
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Discard(tabId int, tabIds ...int) error {
	if response, err := client.Request(&Request{
		Method: "discard",
		TabIds: append(tabIds, tabId),
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Hide(tabId int, tabIds ...int) error {
	if response, err := client.Request(&Request{
		Method: "hide",
		TabIds: append(tabIds, tabId),
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Show(tabId int, tabIds ...int) error {
	if response, err := client.Request(&Request{
		Method: "show",
		TabIds: append(tabIds, tabId),
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) ToggleReaderMode(tabId int) error {
	if response, err := client.Request(&Request{
		Method: "toggleReaderMode",
		TabId: tabId,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) GoBack(tabId int) error {
	if response, err := client.Request(&Request{
		Method: "goBack",
		TabId: tabId,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) GoForward(tabId int) error {
	if response, err := client.Request(&Request{
		Method: "goForward",
		TabId: tabId,
	}); err != nil {
		return err
	} else if response.Status != "success" {
		return fmt.Errorf("Browser responded: %s: %v", response.Status, response.Info)
	}
	return nil
}

func (client *TabsClient) Request(msg *Request) (*Response, error) {
	if client.gatewayConn == nil {
		return nil, errors.New("Cannot send request to closed gateway connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	responseChan := make(chan *Response)
	msg.ID = uuid.New()
	client.requests[msg.ID] = responseChan

	SendMsg(client.gatewayConn, &Message{Request: msg})

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
		switch {
		case msg.Response != nil:
			response := msg.Response
			log.Printf("Received response for %s", response.ID)
			if responseChan, exists := client.requests[response.ID]; exists {
				responseChan <- response
			} else {
				log.Printf("Received unexpected msg response: %v", response)
			}
		case msg.Event != nil:
			client.Updates <- msg.Event
		default:
			log.Printf("Received unexpected msg: %v", msg)
		}
	}
}
