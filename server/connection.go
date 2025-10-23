package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func HandleConnection(conn net.Conn, s *ServerState) {
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
	//format username
	username = strings.TrimSpace(username)
	joinResp := &ServerJoinResponse{}
	s.JoinServer(username, joinResp)

	//print message from the server
	writer.WriteString(joinResp.Message)
	writer.Flush()

	//if banned, end the connection
	if joinResp.Role.Role == RoleBanned {
		return
	} 
	//not banned, allow user to join rooms, starting in no room
	//start listening for input from user
	userInput := make(chan string)
	go getInput(reader, userInput, joinResp.Role.Term)
	
	for {
		select{
		//listen for commands from the server/room
		case msg := <- joinResp.Role.RecvServer:
			//run server/room command
			msg.ExecuteClient()
			
		//listen for input from the user
		case input := <-userInput:
			//convert raw input to metadata (no timestamp)
			rawInput := MsgMetadata{UserName: username, Content: input}
			//send raw data to server
			var reply ExecutableMessage
			s.RecvMessage(&rawInput, &reply)
			//once have response run the .executeClient
			reply.ExecuteClient()
			
		//if user/server is terminated
		case <- joinResp.Role.Term:
			return
		}
	}
}

func getInput(reader *bufio.Reader,userInput chan string, term chan struct{}) {
	for {
		select {
		//if the termination channel is called for a user, terminate reader goroutine
		case <-term:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				close(userInput)
				return
			}
			input := strings.TrimSpace(line)
			userInput <- input
		}
	}
}