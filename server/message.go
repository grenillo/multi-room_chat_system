package server

import (
	"log"
	"multi-room_chat_system/shared"
	"strings"
	"time"
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
		//add metadata to the join type
		join := &JoinCmd{JoinCmd: &shared.JoinCmd{MsgMetadata: input, Room: parts[1]}}
		return join
	case "/leave":
		return &LeaveCmd{LeaveCmd: &shared.LeaveCmd{MsgMetadata: input}}
	case "/listusers":
		return &ListUsersCmd{ListUsersCmd: &shared.ListUsersCmd{MsgMetadata: input, Reply: shared.LUResp{}}}
	case "/help":
		return &HelpCmd{HelpCmd: &shared.HelpCmd{MsgMetadata: input, Invalid: false}}
	case "/quit":
		return &QuitCmd{QuitCmd: &shared.QuitCmd{MsgMetadata: input}}
	case "/kick":
		return &KickBanCmd{KickBanCmd: &shared.KickBanCmd{MsgMetadata: input, Ban: false, User: parts[1]}}
	case "/ban":
		return &KickBanCmd{KickBanCmd: &shared.KickBanCmd{MsgMetadata: input, Ban: true, User: parts[1]}}
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
	//check that user is in a room
	if s.users[m.UserName].CurrentRoom == "" {
		resp = shared.ResponseMD{Status: false, ErrMsg: "PERMISSION DENIED: User is not currently in a room"}
		m.Response = resp
		return
	}
	//user in room, log message
	s.rooms[s.users[m.UserName].CurrentRoom].log = append(s.rooms[s.users[m.UserName].CurrentRoom].log, *m.Message)
	//broadcast to all other users
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
	log.Println(s.rooms)
	log.Println("client attempting to join room:", j.Room)
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

	//broadcast and log user joining to all others currently in the room
	broadcast(j.UserName, "joined", j.Timestamp, j.Room)

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
	l.Room = s.users[l.UserName].CurrentRoom
	remove(l.UserName, l.Room)
	//broadcast and log user leaving to all others currently in the room
	broadcast(l.UserName, "left", l.Timestamp, l.Room)


	//update user state
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
	log.Println("users in room:", s.rooms[s.users[lu.UserName].CurrentRoom])
	lu.Reply.Users = mapToSlice(s.rooms[s.users[lu.UserName].CurrentRoom].users)
	log.Println("log of users:", lu.Reply.Users)
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


////////////////////////////// QUIT CMD and its execute functions ///////////////////////////////
type QuitCmd struct {
	*shared.QuitCmd
}
//user will always be able to quit
func (q *QuitCmd) ExecuteServer() {
	s := GetServerState()
	//check if user is in a room
	if s.users[q.UserName].CurrentRoom != "" {
		room := s.users[q.UserName].CurrentRoom
		remove(q.UserName, room)
		broadcast(q.UserName, "left", q.Timestamp, room)
	}
	//set user status to false
	s.users[q.UserName].Active = false
	//close this connectionHandler once response is sent
	safeClose(s.users[q.UserName].Term)
}
func (q *QuitCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// QUIT CMD and its execute functions ///////////////////////////////
type KickBanCmd struct {
	*shared.KickBanCmd
}
func (kb *KickBanCmd) ExecuteServer() {
	s := GetServerState()
	//first check user's role to see if they can execute
	if (s.users[kb.UserName].Role < RoleAdmin) {
		kb.Status = false
		kb.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	} else { //is an admin or owner
		//check to see if the specified user exists
		if _, exists := s.users[kb.User]; !exists {
			kb.Status = false
			kb.ErrMsg = "PERMISSION DENIED: User: " + kb.User + " does not exist on this server"
			return
		} else { //user exists
			//check if kick and user online
			if !s.users[kb.User].Active && !kb.Ban {
				kb.Status = false
				kb.ErrMsg = "PERMISSION DENIED: User:" + kb.User + " is not logged in"
				return
			}
			//otherwise, check if that user is in a room
			kb.Status = true
			//check if user is in a room
			if s.users[kb.User].CurrentRoom != "" {
				room := s.users[kb.User].CurrentRoom
				remove(kb.User, room)
				broadcast(kb.User, "left", kb.Timestamp, room)
			}
			var msg *Message
			var update *Message
			//if ban
			if kb.Ban {
				s.users[kb.User].Role = RoleBanned
				//broadcast ban to staff
				msg = formatStaffMsg(kb.UserName, "banned user: " + kb.User, kb.Timestamp)
				update = formatStaffMsg("You", "have been banned!", kb.Timestamp)
			} else {
				msg = formatStaffMsg(kb.UserName, "kicked user: " + kb.User, kb.Timestamp)
				update = formatStaffMsg("You", "have been kicked!", kb.Timestamp)
			}

			broadcastToStaff(msg)
			//update user, update its active state, then close its term channel
			s.users[kb.User].RecvServer <- update
			s.users[kb.User].Active = false
			safeClose(s.users[kb.User].Term)
		}
	}
}
func (kb *KickBanCmd) ExecuteClient() {}

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


func broadcast(username string, action string, timestamp time.Time, room string) {
	s := GetServerState()
	//add user action to the room's log
	m := shared.Message{
		MsgMetadata: shared.MsgMetadata{
			Timestamp: timestamp,
			UserName:  username,
			Flag:      true,
			Content:   " " + action + " " + room,
		},
		Response: shared.ResponseMD{Status: true},
	}
	M := &Message{Message: &m}
	s.rooms[room].log = append(s.rooms[room].log, m)

	//broadcast user action to all other users in the room
	s.rooms[room].broadcast(M)
}

func remove(username string, room string) {
	s := GetServerState()
	//remove user from their requested room
	s.rooms[room].removeUser(s.users[username])
	s.users[username].CurrentRoom = ""
}

func formatStaffMsg(username string, action string, timestamp time.Time) *Message{
	m := shared.Message{
		MsgMetadata: shared.MsgMetadata{
			Timestamp: timestamp,
			UserName:  username,
			Flag:      true,
			Content:   " " + action,
		},
		Response: shared.ResponseMD{Status: true},
	}
	return &Message{Message: &m}
}

func broadcastToStaff(msg *Message) {
	s := GetServerState()
	for name, user := range s.users {
		if name == msg.UserName {
			continue
		} 
		if user.Role >= RoleAdmin {
			user.RecvServer <- msg
		}
	}
}