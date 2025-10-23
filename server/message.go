package server

import (
	"fmt"
	"strings"
	"time"
)

//define all methods a message should have
type ExecutableMessage interface {
	ExecuteServer(s *ServerState) //broadcast or command functionality
	ExecuteClient()

}

func MessageFactory(input MsgMetadata, s *ServerState) ExecutableMessage {
	//send to command factory
	if input.Content[0] == '/' {
		return CommandFactory(input, s)
	}
	//otherwise return message
	return &Message{MsgMetadata: input}
}

func CommandFactory (input MsgMetadata, s *ServerState) ExecutableMessage {
	parts := strings.Fields(input.Content)
	switch parts[0] {
	case "/join":
		//get current room for the user
		room := s.users[input.UserName].CurrentRoom
		//add metadata to the join type
		join := &JoinCmd{MsgMetadata: input, Room: room}
		return join
	case "/leave":
		return nil
	case "/listusers":
		return nil
	default:
		return nil
	}
}

type ResponseMD struct {
	Status bool
	ErrMsg string
}

type MsgMetadata struct {
	UserName string
	Timestamp time.Time
	Content string
}

/////////////////////////////// MESSAGE and its execute functions ///////////////////////////////
type Message struct {
	MsgMetadata
	Response ResponseMD
}

func (m *Message) ExecuteServer(s *ServerState) {
	var resp ResponseMD
	//check that user is in this room
	result := contains(mapToSlice(s.users), m.UserName)
	if !result {
		resp = ResponseMD{Status: false, ErrMsg: "PERMISSION DENIED: User is not currently in a room"}
		m.Response = resp
		return
	}
	//user in room, broadcast to all other users
	resp = ResponseMD{Status: true}
	m.Response = resp
	s.rooms[s.users[m.UserName].CurrentRoom].broadcast(m)
}

func (m *Message) ExecuteClient() {
	//check if message was sent
	if !m.Response.Status {
		fmt.Println(m.Response.ErrMsg)
		return
	}
	//otherwise, print to our client's local terminal
	fmt.Println(formatMessage(m))
	
}
///////////////////////////// END MESSAGE and its execute functions /////////////////////////////

////////////////////////////// JOIN CMD and its execute functions //////////////////////////////
type JoinCmd struct {
	MsgMetadata //inherits metadata
	Room string
	Reply JoinResp
}

type JoinResp struct {
	ResponseMD	//inherits
	Log []Message
}

func (j *JoinCmd) ExecuteServer(s *ServerState) {
	//first check that the room exists
	result := contains(mapToSlice(s.rooms), j.Room)
	if !result {
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: Room does not exist"
		return
	}
	//check that the user can join this room
	result = contains(s.users[j.UserName].AvailableRooms, j.Room)
	//if room DNE, set status to false and return
	if !result {
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: Join room request failed"
		return
	}
	//check if the user is already in this room
	if s.users[j.UserName].CurrentRoom == j.Room {
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: User already in specified room"
		return
	}
	//check if the user has permission to join this room
	//banned < member < admin < owner, so if user's role is less than the required permission
	if s.users[j.UserName].Role < s.rooms[j.Room].permission {
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: User role does not have access to room"
		return
	}
	
	//check if user is in a different room, if they are, remove them from that room's state before proceeding
	if _, exists := s.rooms[s.users[j.UserName].CurrentRoom]; exists {
		s.rooms[s.users[j.UserName].CurrentRoom].removeUser(s.users[j.UserName])
	}

	//add user to room
	s.rooms[j.Room].addUser(s.users[j.UserName])
	j.Reply.Status = true
	//update user's room
	s.users[j.UserName].CurrentRoom = j.Room
	//store the room's current state of messages in the response
	j.Reply.Log = s.rooms[j.Room].log

}

func (j *JoinCmd) ExecuteClient() {
	if !j.Reply.Status {
		fmt.Println(j.Reply.ErrMsg)
		return
	}
	clearScreen()
	fmt.Println("=======JOINED ROOM ", j.Room, "=======")
	//print out entire message history to client
	for _, msg := range j.Reply.Log {
		fmt.Println(formatMessage(&msg))
	}
}
////////////////////////// END JOIN CMD and its execute functions //////////////////////////////


////////////////////////////// LEAVE CMD and its execute functions //////////////////////////////
type LeaveCmd struct {
	MsgMetadata
	Room string
	Reply ResponseMD
}

func (l *LeaveCmd) ExecuteServer(s *ServerState) {
	//check to see if the user is in a room
	if _, exists := s.rooms[s.users[l.UserName].CurrentRoom]; !exists {
		l.Reply.Status = false
		l.Reply.ErrMsg = "PERMISSION DENIED: User not in room"
		return
	}
	//remove user from their requested room
	s.rooms[s.users[l.UserName].CurrentRoom].removeUser(s.users[l.UserName])
	l.Room = s.users[l.UserName].CurrentRoom
	//update user state
	s.users[l.UserName].CurrentRoom = ""
	l.Reply.Status = true
}

func (l *LeaveCmd) ExecuteClient() {
	if !l.Reply.Status {
		fmt.Println(l.Reply.ErrMsg)
		return
	}
	clearScreen()
	fmt.Println("=======LEFT ROOM ", l.Room,"=======")	
}
//////////////////////////// END LEAVE CMD and its execute functions ////////////////////////////







//HELPER FUNCTIONS
func contains(container []string, value string) bool {
	for _, v := range container {
		if v == value {
			return true
		}
	}
	return false
}

func mapToSlice[V any](m map[string]V) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func clearScreen() {
    fmt.Print("\033[2J\033[H")
}

func formatMessage(m *Message) string{
	var resp string
	//convert timestamp to string
	time := m.Timestamp.Format("2006-01-02 15:04:05")
	resp = time + "\t" + m.UserName + ":  " + m.Content
	return resp
}