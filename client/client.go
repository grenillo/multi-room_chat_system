package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"multi-room_chat_system/shared"
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
			if msg == nil {
				log.Println("client was terminated, do not forward nil to GUI")
				continue
			}
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


func recvExecutableMsg(term chan struct{}, decoder *gob.Decoder) shared.ExecutableMessage {
	//create new message to return
	var msg interface{}
	err := decoder.Decode(&msg)
	if err != nil {
		if err == io.EOF {
            fmt.Println("Server closed the connection. Exiting...")
        } else {
            fmt.Println("Error decoding message from server:", err)
        }
        close(term) // terminate client if decoding fails
        return nil
	}
	log.Println("client received gob message from server")
	//otherwise return executable to client
	return wrapShared(msg)
}

func wrapShared(msg interface{}) shared.ExecutableMessage {
	switch m := msg.(type) {
    case *shared.HelpCmd:
		return &HelpCmd{HelpCmd: m}
	case *shared.JoinCmd:
		return &JoinCmd{JoinCmd: m}
	case *shared.LeaveCmd:
		return &LeaveCmd{LeaveCmd: m}
	case *shared.ListUsersCmd:
		return &ListUsersCmd{ListUsersCmd: m}
	case *shared.Message:
		return &Message{Message: m}
	case *shared.QuitCmd:
		return &QuitCmd{QuitCmd: m}
	case *shared.KickBanCmd:
		return &KickBanCmd{KickBanCmd: m}
	case *shared.UnBanCmd:
		return &UnBanCmd{UnBanCmd: m}
	case *shared.CreateCmd:
		return &CreateCmd{CreateCmd: m}
	case *shared.DeleteCmd:
		return &DeleteCmd{DeleteCmd: m}
	case *shared.PromoteDemoteCmd:
		return &PromoteDemoteCmd{PromoteDemoteCmd: m}
	case *shared.BroadcastCmd:
		return &BroadcastCmd{BroadcastCmd: m}
	case *shared.ShutdownCmd:
		return &ShutdownCmd{ShutdownCmd: m}
	case *shared.ListRoomsCmd:
		return &ListRoomsCmd{ListRoomsCmd: m}
	case *shared.RoomUpdate:
		return &RoomUpdate{RoomUpdate: m}
	case *shared.UserUpdate:
		return &UserUpdate{UserUpdate: m}
	case *shared.GetLog:
		return &GetLog{GetLog: m}
    default:
        panic("error during wrapping: unknown shared type")
    }
}

