package tabs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
)

/*
 * The Gateway is started by the browser
 * The browser sends messages to the gateway's stdin and reads msg from
 * the gateway's stdout
 */

var (
	GatewaySockAddr = "/tmp/browser_gateway.sock"
	GatewayLogfile = "/home/francis/Projects/firefox-extension/gateway.log"
)

type Gateway struct {
	tabs        *TabStore
	connections []net.Conn
	// receive Response and Events from browser
	inStream    chan *Message
	// send Requests from connections to browser
	outStream   chan *Message
	// since we will be forwarding these onward, send these as generic
	// messages rather than Responses to avoid needless unwrap/rewrap
	requests	map[uuid.UUID]chan *Message
}

func MakeGateway() *Gateway {
	return &Gateway{
		tabs:        MakeTabStore(),
		connections: []net.Conn{},
		requests:	 make(map[uuid.UUID]chan *Message),
		inStream:    make(chan *Message),
		outStream:   make(chan *Message),
	}
}

func (g *Gateway) Start() {
	f, err := os.OpenFile(GatewayLogfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Printf("PID is %d", os.Getpid())

	// send messages from all connections to stdout
	go func(){
		for msg := range g.outStream {
			if err := SendMsg(os.Stdout, msg); err != nil {
				log.Fatalf("Failed to write to stdout (???): %v", err)
			}
		}
	}()

	go func() {
		for msg := range g.inStream {
			if err := g.handleMessage(msg); err != nil {
				log.Printf("ERROR: handling msg: %v", err)
			}
		}
	}()

	go func() {

		stdin := bufio.NewReader(os.Stdin)
		for {
			msg, err := ReadMsg(stdin)
			if err == io.EOF {
				log.Printf("Received EOF from browser")
				os.Exit(0)
			} else if err != nil {
				log.Printf("ERROR: %#v", err)
				continue
			}
			g.inStream <- msg
		}
	}()

	responseChan := make(chan *Message)
	request := &Request{ID: uuid.New(), Method: "list"}
	g.requests[request.ID] = responseChan
	g.outStream <- &Message{Request: request}

	msg := <-responseChan
	delete(g.requests, request.ID)
	if msg.Response == nil || msg.Response.Status != "list" {
		log.Fatalf("Unexpected response to initial query: %v", msg)
	}
	var tabs []*Tab
	if err := json.Unmarshal(msg.Response.Info, &tabs); err != nil {
		log.Fatalf("Unable to read tab list: %v", err)
	}
	log.Printf("Received %d tabs from browser", len(tabs))
	for _, tab := range tabs {
		g.tabs.Open[tab.ID] = tab
	}
	g.listenForConnections()
}

/*
   Handle messages sent by the browser

   If the message id is 0, the message is a push from the browser. The
   change is applied to the gateway's tab store and then broadcast to
   all clients

   Otherwise, it is a response to a request we sent, and we route it
   appropriately
  */
func (g *Gateway) handleMessage(msg *Message) error {
	switch {
	case msg.Request != nil:
		g.outStream <- msg
	case msg.Response != nil:
		response := msg.Response
		if responseChan, exists := g.requests[response.ID]; !exists {
			return fmt.Errorf("Received response for non-outstanding request")
		} else {
			delete(g.requests, response.ID)
			responseChan <- msg
		}
	case msg.Event != nil:
		msg.Event.Apply(g.tabs)
		for _, conn := range g.connections {
			// b, _ := json.Marshal(msg)
			// log.Println("FORWARDING: ", string(b))
			if err := SendMsg(conn, msg); err != nil {
				log.Printf("ERROR: Failed to send msg to %v: %v", conn, err)
			}
		}
	}
	return nil
}

func (g *Gateway) listenForConnections() {
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
		g.connections = append(g.connections, conn)
		log.Println("New client connected")
		go g.listenConn(conn)
	}
}


func (g *Gateway) listenConn(conn net.Conn) {
	defer g.closeConn(conn)

	// I think this works???
	cIn := bufio.NewReader(conn)
	for {
		msg, err := ReadMsg(cIn)
		if err == io.EOF {
			log.Printf("Received EOF from client")
			break
		} else if err != nil {
			log.Printf("ERROR: %#v", err)
			continue
		}

		// The gateway can respond immediately to Read requests
		// and forwards Write requests to the browser
		request := msg.Request
		if request == nil {
			log.Printf("ERROR: Received non-request from client: %v", msg)
			continue
		}
		switch request.Method {
		case "list":
			currentTabs := make([]*Tab, len(g.tabs.Open), len(g.tabs.Open))
			i := 0
			for _, tab := range g.tabs.Open {
				currentTabs[i] = tab
				i++
			}
			if content, err := json.Marshal(currentTabs); err != nil {
				log.Printf("ERROR: Failed to list tabs: %v", err)
			} else {
				response := &Message{Response: &Response{ID: request.ID, Status: "list", Info: content}}
				SendMsg(conn, response)
			}
		default:
			responseChan := make(chan *Message)
			g.requests[request.ID] = responseChan
			go func(resp chan *Message) {
				g.outStream <- msg
				response := <-resp
				SendMsg(conn, response)
			}(responseChan)
		}
	}
}

func (g *Gateway) closeConn(conn net.Conn) {
	log.Printf("Closing connection %v", conn)
	for i, c := range g.connections {
		if conn == c {
			g.connections = append(g.connections[:i], g.connections[i+1:]...)
			break
		}
	}
	conn.Close()
}
