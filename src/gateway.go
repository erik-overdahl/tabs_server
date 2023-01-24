package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
)

type Gateway struct {
	connections map[uint32]net.Conn
	inStream    chan []byte
	outStream   chan []byte
}

func MakeGateway() *Gateway {
	return &Gateway{
		connections: map[uint32]net.Conn{},
		inStream:    make(chan []byte),
		outStream:   make(chan []byte),
	}
}

func (g *Gateway) Start() {
	f, err := os.OpenFile("/home/francis/Projects/firefox-extension/gateway.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Printf("PID is %d", os.Getpid())

	go g.sendBrowserMsgs()

	// send messages from all connections to stdout
	go g.passMsgsToBrowser()

	go g.listenForConnections()

	stdin := bufio.NewReader(os.Stdin)
	for {
		buf, err := ReadBrowserMsg(stdin)
		if err == io.EOF {
			log.Printf("Received EOF from browser")
			break
		} else if err != nil {
			log.Printf("ERROR: %#v", err)
			continue
		}
		// log.Printf("Got msg of len %d bytes", len(buf))

		g.inStream <- buf
	}
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
		id := rand.Uint32()

		log.Printf("New connection accepted (id %d)", id)
		_, err = SendBrowserMsg(conn, []byte(fmt.Sprintf(`{"action":"clientId", "content": %d}`, id)))
		if err != nil {
			log.Printf("ERROR: send clientId msg to client: %v", err)
		}
		g.connections[id] = conn
		go g.listenConn(conn)
	}

}

func (g *Gateway) sendBrowserMsgs() {
	for msg := range g.inStream {
		var body map[string]json.RawMessage
		if err := json.Unmarshal(msg[4:], &body); err != nil {
			log.Printf("ERROR: Failed to parse message: %v", err)
			continue
		}
		if clientIdRaw, exists := body["clientId"]; exists {
			var clientId uint32
			if err := json.Unmarshal(clientIdRaw, &clientId); err != nil {
				log.Printf("ERROR: Failed to parse client id: %v", err)
				continue
			}
			if client, exists := g.connections[clientId]; exists {
				client.Write(msg)
			} else {
				log.Printf("ERROR: Received message for unknown client %d", clientId)
			}
		} else {
			for clientId, client := range g.connections {
				_, err := client.Write(msg)
				if err != nil {
					log.Printf("ERROR: writing to client %d: %v", clientId, err)
				}
			}
		}
	}
}

func (g *Gateway) listenConn(conn net.Conn) {
	defer g.closeConn(conn)

	// I think this works???
	cIn := bufio.NewReader(conn)
	for {
		buf, err := ReadBrowserMsg(cIn)
		if err == io.EOF {
			log.Printf("Received EOF from client")
			break
		} else if err != nil {
			log.Printf("ERROR: %#v", err)
			continue
		}
		// log.Printf("Got msg of len %d bytes from client", len(buf))

		g.outStream <- buf
	}
}

func (g *Gateway) passMsgsToBrowser() {
	for msg := range g.outStream {
		_, err := SendBrowserMsg(os.Stdout, msg)
		if err != nil {
			log.Fatalf("Failed to write to stdout (???): %v", err)
		}
	}
}

func (g *Gateway) closeConn(conn net.Conn) {
	log.Printf("Closing connection")
	for clientId, client := range g.connections {
		if conn == client {
			delete(g.connections, clientId)
			break
		}
	}
	conn.Close()
}
