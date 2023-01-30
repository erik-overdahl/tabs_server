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
	requests	map[uuid.UUID]net.Conn
	inStream    chan *Message
	outStream   chan *Message
}

func MakeGateway() *Gateway {
	return &Gateway{
		tabs:        MakeTabStore(),
		connections: []net.Conn{},
		requests:	 map[uuid.UUID]net.Conn{},
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
			if err := SendMsg(os.Stdout, *msg); err != nil {
				log.Fatalf("Failed to write to stdout (???): %v", err)
			}
		}
	}()

	go func() {
		for msg := range g.inStream {
			if err := g.handleMessage(*msg); err != nil {
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
				break
			} else if err != nil {
				log.Printf("ERROR: %#v", err)
				continue
			}
			g.inStream <- msg
		}
	}()

	g.outStream <- &Message{ID: uuid.Nil, Action: "list"}
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
func (g *Gateway) handleMessage(msg Message) error {
	if msg.ID != uuid.Nil {
		log.Printf("Msg for %s", msg.ID)
		if conn, exists := g.requests[msg.ID]; !exists {
			return fmt.Errorf("Received response for non-outstanding request")
		} else {
			log.Println("Sending to", conn)
			delete(g.requests, msg.ID)
			return SendMsg(conn, msg)
		}
	}
	if err := g.tabs.Apply(&msg); err != nil {
		return err
	}
	for _, conn := range g.connections {
		log.Println("FORWARDING: ", conn)
		if err := SendMsg(conn, msg); err != nil {
			log.Printf("Failed to send msg to %v: %v", conn, err)
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
		switch msg.Action {
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
				response := Message{ID: msg.ID, Action: msg.Action, Content: content}
				SendMsg(conn, response)
			}
		default:
			g.requests[msg.ID] = conn
			g.outStream <- msg
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
