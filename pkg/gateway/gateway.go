package gateway

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
)

var (
	GatewaySockAddr string
)

// if an error occurs when handling a request

type RequestHandler interface {
	Handle(*Request) *Response
}

type EventHandler interface {
	Handle(*Event) *Event
}

// handles the "subscribe" method

type subscriptionHandler struct {
	// these maps are shared with the Gateway
	// this may be indicative of a poor design
	listeners map[string][]net.Conn
	requests map[uuid.UUID]net.Conn
}

func (this *subscriptionHandler) Handle(req *Request) *Response {
	eventName, ok := req.Data.(string)
	if !ok {
		return &Response{
			Status: STATUS_FAILURE,
			Data:   []byte(fmt.Sprintf("failed to read event name: expected %T, got %T: %v", "", req.Data, req.Data)),
		}
	}
	listener := this.requests[req.Id]
	if listeners, exists := this.listeners[eventName]; !exists {
		this.listeners[eventName] = []net.Conn{listener}
	} else {
		this.listeners[eventName] = append(listeners, listener)
	}
	return &Response{Id: req.Id, Status: STATUS_SUCCESS}
}

type Gateway struct {
	// requests coming in from clients
	inStream    chan *Message
	requests    map[uuid.UUID]net.Conn
	listeners   map[string][]net.Conn

	eventHandlers   map[string]EventHandler
	requestHandlers map[string]RequestHandler
}

func MakeGateway() *Gateway {
	gateway := &Gateway{
		inStream:      make(chan *Message),
		requests:      make(map[uuid.UUID]net.Conn),
		listeners: 	   make(map[string][]net.Conn),
		eventHandlers: make(map[string]EventHandler),
	}
	gateway.requestHandlers = map[string]RequestHandler{
		"subscribe": &subscriptionHandler{
			requests: gateway.requests,
			listeners: gateway.listeners,
		},
	}
	return gateway
}

func (this *Gateway) RegisterRequestHandler(method string, handler RequestHandler) error {
	if _, exists := this.requestHandlers[method]; exists {
		return fmt.Errorf("method '%s' already has a handler registered", method)
	}
	this.requestHandlers[method] = handler
	return nil
}

func (this *Gateway) RegisterEventHandler(event string, handler EventHandler) error {
	if _, exists := this.eventHandlers[event]; exists {
		return fmt.Errorf("event '%s' already has a handler registered", event)
	}
	this.eventHandlers[event] = handler
	return nil
}

// I don't check to see if it's already running. Just everyone agree to
// only call this once, ok?
func (this *Gateway) Run() error {
	go this.sendToBrowser()
	go this.listenForConnections()
	// read messages from the browser
	stdin := bufio.NewReader(os.Stdin)
	for {
		msg, err := ReadMsg(stdin)
		if errors.Is(err, io.EOF) {
			log.Println("EOF from browser, shutting down")
			os.Exit(0)
		} else if err != nil {

		}
		this.handleMessage(msg)
	}
}

func (this *Gateway) handleMessage(msg *Message) error {
	switch {
	// received from clients
	case msg.Request != nil:
		if handler, exists := this.requestHandlers[msg.Request.Method]; exists {
			response := handler.Handle(msg.Request)
			conn := this.requests[msg.Request.Id]
			return SendMsg(conn, &Message{Response: response})
		}
		this.inStream <- msg

	// received from the browser
	case msg.Response != nil:
		conn, exists := this.requests[msg.Response.Id];
		if !exists {
			return fmt.Errorf("Received response for unknown request: %s", msg.Response.Id)
		}
		return SendMsg(conn, msg)

	case msg.Event != nil:
		event := msg.Event
		if handler, exists := this.eventHandlers[event.Type]; exists {
			event = handler.Handle(event)
		}
		return this.broadcast(event)
	}
	return nil
}

func (this *Gateway) broadcast(event *Event) error {
	msg := &Message{Event: event}
	listeners, exists := this.listeners[event.Type]
	if !exists {
		return nil
	}
	for _, listener := range listeners {
		if err := SendMsg(listener, msg); err != nil {

		}
	}
	return nil
}

func (this *Gateway) sendToBrowser() {
	for msg := range this.inStream {
		if err := SendMsg(os.Stdout, msg); err != nil {
			log.Fatalf("Failed to write to stdout (???): %v", err)
		}
	}
}

func (this *Gateway) listenForConnections() {
	if err := os.RemoveAll(GatewaySockAddr); err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("unix", GatewaySockAddr)
	if err != nil {
		log.Fatal("ERROR: listening: ", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("ERROR: accept: ", err)
		}
		go this.handleConnection(conn)
	}
}

func (this *Gateway) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		msg, err := ReadMsg(conn)

		if errors.Is(err, io.EOF) {
			return // the client hung up
		} else if err != nil {
			// what do we do here? should we try to inform the client
			// that we couldn't read a message?
		} else if msg.Request == nil {
			log.Printf("ERROR: Received non-request from client: %v", msg)
			continue
		}

		this.requests[msg.Request.Id] = conn
		this.handleMessage(msg)
	}
}
