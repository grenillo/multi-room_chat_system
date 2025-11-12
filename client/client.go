package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"multi-room_chat_system/shared"
	"net"
	"os"
	"strings"
)

func StartClient() {
	//register types with gob for tcp
	shared.Init()
	//connect to the server
	conn, err := net.Dial("tcp", "localhost:5461")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected to server:", conn.RemoteAddr())

	//channl for termination
	term := make(chan struct{})
	fmt.Println("Term chan created")

	//read initial username prompt from server
	serverReader := bufio.NewReader(conn)
	prompt, _ := serverReader.ReadString('>')
	prompt = strings.TrimSuffix(prompt, ">")
	fmt.Print(prompt)
	var username string
	fmt.Scanln(&username)
	//send to server
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		fmt.Println("Failed to send username:", err)
		close(term)
		return
	}
	//read response
	resp, _ := serverReader.ReadString('>')
	resp = strings.TrimSuffix(resp, ">")
	if strings.Contains(resp, "PERMISSION DENIED") {
		fmt.Println(resp)
		close(term)
		return
	}
	ClearScreen()
	fmt.Println(resp)
	//create goroutines to continuously listen for input and server responses
	go getInput(conn, term)
	//fmt.Println("input stream created")
	go outputFromServer(conn, term) // ensures recvExecutableMsg is used
	//fmt.Println("output stream created")

	//block main from immediately exiting
	//fmt.Println("blocked from exiting")
	<-term
	//fmt.Println("unblocked from exiting")

}

//read user input and send it to the server
func getInput(conn net.Conn,  term chan struct{}) {
	//create reader for this client
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			close(term)
			return
		}
		input := strings.TrimSpace(line)
		fmt.Println("sending to server input: ", input)
		fmt.Fprintln(conn, input)
	}
}

func outputFromServer(conn net.Conn, term chan struct{}) {
	decoder := gob.NewDecoder(conn)
	for {
		select {
		case <-term:
			return
		default:
			//receive and decode executable message from server
			msg := recvExecutableMsg(term, decoder)
			if msg == nil {
                return
            }
			//execute client-side logic sent from server state
			msg.ExecuteClient()
		}	
	}
}

func recvExecutableMsg(term chan struct{}, decoder *gob.Decoder) shared.ExecutableMessage{
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
	case *shared.CreateCmd:
		return &CreateCmd{CreateCmd: m}
	case *shared.DeleteCmd:
		return &DeleteCmd{DeleteCmd: m}
    default:
        panic("error during wrapping: unknown shared type")
    }
}