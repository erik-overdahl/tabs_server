package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	cmd := "serve"
	if len(os.Args) > 0 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "gateway":
		gateway := MakeGateway()
		gateway.Start()
	case "serve":
		server := MakeTabsServer()
		err := server.ConnectBrowserGateway()
		if err != nil {
			log.Fatalf("Unable to connect to browser gateway: %v", err)
		}
		server.ReadBrowserMsgs()
	case "client":
	default:
		log.Printf("ERROR: passed unknown command '%s'", cmd)
		os.Exit(1)
	}
}

func ReadBrowserMsg(r io.Reader) ([]byte, error) {
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
		return buf, fmt.Errorf("ERROR: reading msg: %v", err)
	}
	return buf, nil
}

func SendBrowserMsg(w io.Writer, buf []byte) (int, error) {
	// add size if not already there
	if len(buf) > 5 && buf[4] != '{' {
		size := make([]byte, 4)
		binary.LittleEndian.PutUint32(size, uint32(len(buf)))
		buf = append(size, buf...)
	}
	return w.Write(buf)
}
