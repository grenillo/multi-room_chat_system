package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
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
	writer.WriteString("Enter your username: >")
	writer.Flush()
	//read the entered username
	username, err := reader.ReadString('\n')
	username = strings.TrimSpace(username)
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
			log.Println("client connectionHandler reveived:", input)
			//convert raw input to metadata (no timestamp)
			rawInput := shared.MsgMetadata{UserName: user.Username, Content: input}
			//send raw data to server
			var reply shared.ExecutableMessage
			log.Println("client connectionHandler sent to server:", input)
			s.RecvMessage(&rawInput, &reply)
			log.Println("client connectionHandler recv response from server for :", input)

			//once have response, forward to client
			forwardToClient(encoder, reply)
			
		//if user/server is terminated
		case <-user.Term:
			close(user.RecvServer)
			close(user.ToServer)
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
				safeClose(user.Term)
				return
			}
			//format input
			input := strings.TrimSpace(line)
			log.Println("received client input:",input)
			//send input to handleConnection
			userInput <- input
		}
	}
}

func forwardToClient(encoder *gob.Encoder, msg shared.ExecutableMessage) error {
	//unwrap the server to the shared type to send to client
	concrete := unwrapShared(msg)
	err := encoder.Encode(&concrete)
	if err != nil {
        fmt.Println("Error sending ExecutableMessage:", err)
        return err
    }
    return nil
}

func unwrapShared(msg interface{}) interface{} {
    switch m := msg.(type) {
    case *HelpCmd:
        return m.HelpCmd		// *shared.HelpCmd
    case *JoinCmd:
        return m.JoinCmd		// *shared.JoinCmd
    case *LeaveCmd:
        return m.LeaveCmd		// *shared.LeaveCmd
    case *ListUsersCmd:
        return m.ListUsersCmd	// *shared.ListUsersCmd
    case *Message:
        return m.Message		// *shared.Message
	case *QuitCmd:
		return m.QuitCmd		// *shared.QuitCMD
    default:
        panic("error during unwrapping: unknown command type")
    }
}


func safeClose(ch chan struct{}) {
    select {
    case <-ch:
        // already closed, do nothing
    default:
        close(ch)
    }
}