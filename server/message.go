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
		return &LeaveCmd{MsgMetadata: input}
	case "/listusers":
		return &ListUsersCmd{MsgMetadata: input}
	case "/help":
		return &HelpCmd{MsgMetadata: input, Invalid: false}
	default:
		return &HelpCmd{MsgMetadata: input, Invalid: true}
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
/////////////////////////////////////////////////////////////////////////////////////////////////


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
/////////////////////////////////////////////////////////////////////////////////////////////////


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
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// LISTUSERS CMD and its execute functions ////////////////////////////
type ListUsersCmd struct {
	MsgMetadata
	Reply LUResp
}
type LUResp struct {
	ResponseMD
	Room string
	Users []string
}
func (lu *ListUsersCmd) ExecuteServer(s* ServerState) {
	//check to see if a user is in a room
	if _, exists := s.rooms[s.users[lu.UserName].CurrentRoom]; !exists {
		lu.Reply.Status = false
		lu.Reply.ErrMsg = "PERMISSION DENIED: User not in room"
		return
	}
	//get the list of users from a room
	lu.Reply.Users = mapToSlice(s.rooms[lu.UserName].users)
	lu.Reply.Status = true
	lu.Reply.Room = s.users[lu.UserName].CurrentRoom
}

func (lu *ListUsersCmd) ExecuteClient() {
	if !lu.Reply.Status {
		fmt.Println(lu.Reply.ErrMsg)
		return
	}
	//if successful print the current room and its users to the client
	fmt.Println("=======CURRENT USERS IN ROOM ",lu.Reply.Room,"=======")
	for _, user := range lu.Reply.Users {
		fmt.Println(user)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// HELP CMD and its execute functions ///////////////////////////////
type HelpCmd struct {
	MsgMetadata
	Invalid bool
	Reply HelpResp
}
type HelpResp struct {
	ResponseMD
	Usage []string
}
func (h *HelpCmd) ExecuteServer(s* ServerState) {
	if h.Invalid {
		h.Reply.Status = false
		h.Reply.ErrMsg = "PERMISSION DENIED: Invalid command, enter /help for more information"
	}
	//get the usage for this user's role
	h.Reply.Status = true
	h.Reply.Usage = getUsage(s.users[h.UserName].Role)
}
func (h *HelpCmd) ExecuteClient() {
	if h.Invalid {
		fmt.Println(h.Reply.ErrMsg)
		return
	}
	fmt.Println("=======USER COMMAND USAGE=======")	
	//print the commands for this user
	for _, cmd := range h.Reply.Usage {
		fmt.Println(cmd)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////


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