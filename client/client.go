package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"multi-room_chat_system/shared"
	"net"
	"os"
	"strings"
)

func main() {
	//connect to the server
	conn, err := net.Dial("tcp", "localhost:4561")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected to server:", conn.RemoteAddr())

	//channl for termination
	term := make(chan struct{})

	//create goroutines to continuously listen for input and server responses
	go getInput(conn, term)
	go outputFromServer(conn, term) // ensures recvExecutableMsg is used

	//block main from immediately exiting
	<-term

}

//read user input and send it to the server
func getInput(conn net.Conn,  term chan struct{}) {
	//create reader for this client
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ") // indicate prompt 
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			close(term)
			return
		}
		input := strings.TrimSpace(line)
		fmt.Fprintln(conn, input)
	}
}

func outputFromServer(conn net.Conn,  term chan struct{}) {
	for {
		select {
		case <-term:
			return
		default:
			//receive and decode executable message from server
			msg := recvExecutableMsg(conn, term)
			if msg == nil {
                return
            }
			//execute client-side logic sent from server state
			msg.ExecuteClient()
		}	
	}
}

func recvExecutableMsg(conn net.Conn,  term chan struct{}) shared.ExecutableMessage{
	decoder := gob.NewDecoder(conn)
	//create new message to return
	var msg shared.ExecutableMessage
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Error decoding message from server:", err)
        close(term) // terminate client if decoding fails
        return nil
	}
	//otherwise return executable to client
	return msg
}