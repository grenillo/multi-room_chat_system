package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"multi-room_chat_system/shared"
	"net"
	"strings"
)

//function to synchronously check a user's connection before allowing them to input commands to the server
func handleNewConnection(conn net.Conn) {
	//setup new reader and writer
	reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)
	//prompt user for username
	writer.WriteString("Enter your username: ")
	writer.Flush()
	//read the entered username
	username, err := reader.ReadString('\n')
	if err != nil {
        fmt.Println("Failed to read username:", err)
        return
    }
	//get server state
	s := GetServerState()
	//send JoinRPC to the server state
	resp := &ServerJoinResponse{}
	s.JoinServer(username, resp)
	//if user is banned
	if !resp.Status {
		writer.WriteString(resp.Message)
        writer.Flush()
        conn.Close()
        return
	}
	//send confirmation
	writer.WriteString(resp.Message)
    writer.Flush()
	//otherwise start goroutine to handle client requests
	go handleConnection(reader, conn, resp.Role)
	
	
}

//function to asynchronously handle connections once they are verified
func handleConnection(reader *bufio.Reader, conn net.Conn, user *Member) {
	//start listener goroutine to listen for user input
	userInput := make(chan string)
	go getUserInput(reader, user, userInput)
	//get server state (for RPCs)
	s := GetServerState()
	//create encoder for gob usage
	encoder := gob.NewEncoder(conn)
	//start listener goroutine to listen for user input

	for {
		select{
		//listen for commands from the server/room
		case msg := <-user.RecvServer:
			//send client a response from the server
			forwardToClient(encoder, msg)

		//listen for input from the user
		case input := <-userInput:
			//convert raw input to metadata (no timestamp)
			rawInput := shared.MsgMetadata{UserName: user.Username, Content: input}
			//send raw data to server
			var reply shared.ExecutableMessage
			s.RecvMessage(&rawInput, &reply)
			//once have response, forward to client
			forwardToClient(encoder, reply)
			
		//if user/server is terminated
		case <-user.Term:
			conn.Close()
			return
		}
	}

}

func getUserInput(reader *bufio.Reader, user *Member, userInput chan string) {
	for {
		select {
		//if the termination channel is called for a user, terminate reader goroutine
		case <-user.Term:
			return
		default:
			//read line from client
			line, err := reader.ReadString('\n')
			//detect if client disconnects
			if err != nil {
				fmt.Println("Client disconnected")
				close(user.Term)
				return
			}
			//format input
			input := strings.TrimSpace(line)
			//send input to handleConnection
			userInput <- input
		}
	}
}

func forwardToClient(encoder *gob.Encoder, msg shared.ExecutableMessage) error {
	err := encoder.Encode(msg)
	if err != nil {
        fmt.Println("Error sending ExecutableMessage:", err)
        return err
    }
    return nil
}