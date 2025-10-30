package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

//function to synchronously check a user's connection before allowing them to input commands to the server
func handleNewConnection(conn net.Conn) {
	//setup new reader and writer
	reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)
	//prompt user for username
	writer.WriteString("Enter your username: \n>")
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
	//otherwise start goroutine to handle client requests
	go handleConnection(conn, resp.Role, reader)
	
	
}

//function to asynchronously handle connections once they are verified
func handleConnection(conn net.Conn, user *Member, reader *bufio.Reader) {
	//get server state (for RPCs)
	s := GetServerState()
	//start listener goroutine to listen for user input
	userInput := make(chan string)
	go getUserInput(reader, user, userInput)

	for {
		select{
		//listen for commands from the server/room
		case msg := <-user.RecvServer:
			//run server/room command
			msg.ExecuteClient()
			
		//listen for input from the user
		case input := <-userInput:
			//convert raw input to metadata (no timestamp)
			rawInput := MsgMetadata{UserName: user.Username, Content: input}
			//send raw data to server
			var reply ExecutableMessage
			s.RecvMessage(&rawInput, &reply)
			//once have response run the .executeClient
			reply.ExecuteClient()
			
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