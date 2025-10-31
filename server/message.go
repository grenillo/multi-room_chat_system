package server

import (
	"fmt"
	"strings"
	"multi-room_chat_system/shared"
)

func MessageFactory(input shared.MsgMetadata, s *ServerState) shared.ExecutableMessage {
	//send to command factory
	if input.Content[0] == '/' {
		return CommandFactory(input, s)
	}
	//otherwise return message
	return &Message{Message: &shared.Message{MsgMetadata: input}}
}

func CommandFactory (input shared.MsgMetadata, s *ServerState) shared.ExecutableMessage {
	parts := strings.Fields(input.Content)
	switch parts[0] {
	case "/join":
		//get current room for the user
		room := s.users[input.UserName].CurrentRoom
		//add metadata to the join type
		join := &JoinCmd{JoinCmd: &shared.JoinCmd{MsgMetadata: input, Room: room}}
		return join
	case "/leave":
		return &LeaveCmd{LeaveCmd: &shared.LeaveCmd{MsgMetadata: input}}
	case "/listusers":
		return &ListUsersCmd{ListUsersCmd: &shared.ListUsersCmd{MsgMetadata: input}}
	case "/help":
		return &HelpCmd{HelpCmd: &shared.HelpCmd{MsgMetadata: input, Invalid: false}}
	default:
		return &HelpCmd{HelpCmd: &shared.HelpCmd{MsgMetadata: input, Invalid: true}}
	}
}

/////////////////////////////// MESSAGE and its execute functions ///////////////////////////////

type Message struct {
	*shared.Message
}

func (m *Message) ExecuteServer() {
	s := GetServerState()
	var resp shared.ResponseMD
	//check that user is in this room
	result := contains(mapToSlice(s.users), m.UserName)
	if !result {
		resp = shared.ResponseMD{Status: false, ErrMsg: "PERMISSION DENIED: User is not currently in a room"}
		m.Response = resp
		return
	}
	//user in room, broadcast to all other users
	resp = shared.ResponseMD{Status: true}
	m.Response = resp
	s.rooms[s.users[m.UserName].CurrentRoom].broadcast(m)
}

func (m *Message) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


////////////////////////////// JOIN CMD and its execute functions //////////////////////////////
type JoinCmd struct {
	*shared.JoinCmd
}

func (j *JoinCmd) ExecuteServer() {
	s := GetServerState()
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
func (j *JoinCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


////////////////////////////// LEAVE CMD and its execute functions //////////////////////////////
type LeaveCmd struct {
	*shared.LeaveCmd
}

func (l *LeaveCmd) ExecuteServer() {
	s := GetServerState()
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

func (l *LeaveCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// LISTUSERS CMD and its execute functions ////////////////////////////
type ListUsersCmd struct {
	*shared.ListUsersCmd
}

func (lu *ListUsersCmd) ExecuteServer() {
	s := GetServerState()
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

func (lu *ListUsersCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// HELP CMD and its execute functions ///////////////////////////////
type HelpCmd struct {
	*shared.HelpCmd
}

func (h *HelpCmd) ExecuteServer() {
	s := GetServerState()
	if h.Invalid {
		h.Reply.Status = false
		h.Reply.ErrMsg = "PERMISSION DENIED: Invalid command, enter /help for more information"
	}
	//get the usage for this user's role
	h.Reply.Status = true
	h.Reply.Usage = getUsage(s.users[h.UserName].Role)
}
func (h *HelpCmd) ExecuteClient() {}
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
