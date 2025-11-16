package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"multi-room_chat_system/shared"
	"net"
	"strings"
)

type ClientAdapter struct {
    Conn     net.Conn
    Incoming chan shared.ExecutableMessage
    Outgoing chan string
    Term     chan struct{}
	Encoder *gob.Encoder
	Decoder *gob.Decoder
}

func ConnectToServer(username string) (*ClientAdapter, string,  error){
	//register gob
	shared.Init()
	//connect to the server
	conn, err := net.Dial("tcp", "localhost:5461")
    if err != nil {
        return nil, "", fmt.Errorf("could not connect: %w", err)
    }
	//client adapter type
	adapter := &ClientAdapter{
        Conn:      conn,
        Encoder:   gob.NewEncoder(conn),
        Decoder:   gob.NewDecoder(conn),
        Incoming:  make(chan shared.ExecutableMessage),
        Outgoing:  make(chan string),
        Term:	   make(chan struct{}),
    }

	reader := bufio.NewReader(conn)
	//server requests username
	_, err = reader.ReadString('>')
    if err != nil {
        return nil, "", fmt.Errorf("login prompt read failed: %w", err)
    }
	// send username
    _, err = conn.Write([]byte(username + "\n"))
    if err != nil {
        return nil, "", fmt.Errorf("failed sending username: %w", err)
    }
	// read login response
    resp, err := reader.ReadString('>')
	resp = strings.TrimSuffix(resp, "\n>")
    if err != nil {
        return nil, "", fmt.Errorf("failed reading login response: %w", err)
    }
	log.Println(resp)
	//if user is banned
    if strings.Contains(resp, "PERMISSION DENIED"){
        return nil, resp, nil
    }
	//start goroutines to read/write from the GUI
	go adapter.readLoop()
    go adapter.writeLoop()

	return adapter, resp, nil
}

//goroutine to send response to GUI to display for client
func (c * ClientAdapter) readLoop() {
	for {
        select {
        case <-c.Term:
            return
        default:
            msg := recvExecutableMsg(c.Term, c.Decoder)
			log.Println("client received wrapped type")
            //err := c.Decoder.Decode(&msg)
			log.Println("Decoded message:", msg)
            c.Incoming <- msg
        }
    }
}

//writer goroutine
func (c * ClientAdapter) writeLoop() {
	for {
        select {
        case <-c.Term:
            return
        case outgoing := <-c.Outgoing:
            _, err := c.Conn.Write([]byte(outgoing + "\n"))
            if err != nil {
                close(c.Term)
                return
            }
        }
    }
}