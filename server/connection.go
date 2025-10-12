package server

import(
	"net"
	"bufio"
	"fmt"
	"strings"
	"multi-room_chat_system/common"
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
	joinResp := &common.JoinResponse{}
	//send JoinServer RPC to server
	s.JoinServer(username, joinResp)

	//print message from the server
	writer.WriteString(joinResp.Message)
	writer.Flush()

	//if banned, end the connection
	if joinResp.Role == common.RoleBanned {
		return
	} 
	//not banned, allow user to join rooms, starting in no room
	for {
		
	}
}