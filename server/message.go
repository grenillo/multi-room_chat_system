package server

import (
	"log"
	"multi-room_chat_system/shared"
	"strings"
	"time"
)

func MessageFactory(input shared.MsgMetadata, s *ServerState) shared.ExecutableMessage {
	//error check
	if len(input.Content) == 0 {
		return &HelpCmd{HelpCmd: &shared.HelpCmd{MsgMetadata: input, Invalid: true}}
	}
	//send to command factory
	if input.Content[0] == '/' {
		return CommandFactory(input, s)
	}
	//otherwise return message
	return &Message{Message: &shared.Message{MsgMetadata: input}}
}

func CommandFactory (input shared.MsgMetadata, s *ServerState) shared.ExecutableMessage {
	parts := strings.Fields(input.Content)
	//set the args part of the metadata
	input.Args = len(parts)
	switch parts[0] {
	case "/join":
		return &JoinCmd{JoinCmd: &shared.JoinCmd{MsgMetadata: input}}
	case "/leave":
		return &LeaveCmd{LeaveCmd: &shared.LeaveCmd{MsgMetadata: input}}
	case "/listusers":
		return &ListUsersCmd{ListUsersCmd: &shared.ListUsersCmd{MsgMetadata: input, Reply: shared.LUResp{}}}
	case "/help":
		return &HelpCmd{HelpCmd: &shared.HelpCmd{MsgMetadata: input, Invalid: false}}
	case "/quit":
		return &QuitCmd{QuitCmd: &shared.QuitCmd{MsgMetadata: input}}
	case "/kick":
		return &KickBanCmd{KickBanCmd: &shared.KickBanCmd{MsgMetadata: input, Ban: false, }}
	case "/ban":
		return &KickBanCmd{KickBanCmd: &shared.KickBanCmd{MsgMetadata: input, Ban: true, }}
	case "/create":
		return &CreateCmd{CreateCmd: &shared.CreateCmd{MsgMetadata: input}}
	case "/delete":
		return &DeleteCmd{DeleteCmd: &shared.DeleteCmd{MsgMetadata: input}}
	case "/promote":
		return &PromoteDemoteCmd{PromoteDemoteCmd: &shared.PromoteDemoteCmd{MsgMetadata: input, Promote: true}}
	case "/demote":
		return &PromoteDemoteCmd{PromoteDemoteCmd: &shared.PromoteDemoteCmd{MsgMetadata: input, Promote: false}}
	case "/broadcast":
		parts = strings.SplitN(input.Content, " ", 2)
		input.Args = len(parts)
		return &BroadcastCmd{BroadcastCmd: &shared.BroadcastCmd{MsgMetadata: input}}
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
	//check that the cmd was entered properly
	if j.Args != 2 {
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//set the room
	parts := strings.Fields(j.Content)
	j.Room = parts[1]
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
	if s.users[j.UserName].CurrentRoom != "" && j.Room != s.users[j.UserName].CurrentRoom {
		//broadcast user leaving to other members in that room
		broadcast(j.UserName, "left", j.Timestamp, s.users[j.UserName].CurrentRoom)
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
	//check that the cmd was entered properly
	if l.Args != 1 {
		l.Reply.Status = false
		l.Reply.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
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
	//check that the cmd was entered properly
	if lu.Args != 2 {
		lu.Reply.Status = false
		lu.Reply.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
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

//////////////////////////// KICK/BAN CMD and its execute functions /////////////////////////////
type KickBanCmd struct {
	*shared.KickBanCmd
}
func (kb *KickBanCmd) ExecuteServer() {
	s := GetServerState()
	//check that the cmd was entered properly
	if kb.Args != 2 {
		kb.Status = false
		kb.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//set user after verifying it exists
	parts := strings.Fields(kb.Content)
	kb.User = parts[1]
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


///////////////////////////// CREATE CMD and its execute functions //////////////////////////////
type CreateCmd struct {
	*shared.CreateCmd
}
func (c *CreateCmd) ExecuteServer() {
	s := GetServerState()
	//check that the cmd was entered properly
	if c.Args != 3 {
		c.Status = false
		c.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//set room and role after verifying they exist
	parts := strings.Fields(c.Content)
	c.Room = parts[1]
	c.Role = int(convToRole(parts[2]))

	//first check user's role to see if they can execute
	if (s.users[c.UserName].Role < RoleAdmin) {
		c.Status = false
		c.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	}
	//check to see if the room already exists
	if _, exists := s.rooms[c.Room]; exists {
		c.Status = false
		c.ErrMsg = "PERMISSION DENIED: Room already exists"
		return
	}
	//check that the room name is in the correct format
	if c.Room[0] != '#' {
		c.Status = false
		c.ErrMsg = "PERMISSION DENIED: New room name must begin with '#'"
		return
	}
	//check that the user entered the correct permission
	if c.Role == int(RoleBanned) {
		c.Status = false
		c.ErrMsg = "PERMISSION DENIED: Incorrect permission: must be 'all' or 'staff'"
		return
	}

	//initialize the room's state
	newRoom := Room{
		users: make(map[string]*Member),
		log: make([]shared.Message, 0),
		permission: Role(c.Role),
	}
	//add new room to the server's state
	s.rooms[c.Room] = &newRoom
	//update user states
	for _, user := range s.users {
		//only update if they are at least the correct role
		if user.Role >= newRoom.permission {
			user.AvailableRooms = append(user.AvailableRooms, c.Room)
		}
	}
	c.Status = true
	c.ErrMsg = "SERVER: room " + c.Room + " was successfully created"
}
func (c *CreateCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// DELETE CMD and its execute functions //////////////////////////////
type DeleteCmd struct {
	*shared.DeleteCmd
}
func (d *DeleteCmd) ExecuteServer() {
	s := GetServerState()
	//check that the cmd was entered properly
	if d.Args != 2 {
		d.Status = false
		d.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//set room and role after verifying they exist
	parts := strings.Fields(d.Content)
	d.Room = parts[1]
	//first check user's role to see if they can execute
	if (s.users[d.UserName].Role < RoleAdmin) {
		d.Status = false
		d.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	}
	//check to see if the specified room exists
	if _, exists := s.rooms[d.Room]; !exists {
		d.Status = false
		d.ErrMsg = "PERMISSION DENIED: Room " + d.Room + " does not exist" 
		return
	}
	//once here room exists and user has the correct permission
	if d.Room == s.users[d.UserName].CurrentRoom {
		d.InRoom = true
	}
	//generate special leave cmd to send to users
	force := &LeaveCmd{LeaveCmd: &shared.LeaveCmd{ MsgMetadata: shared.MsgMetadata{ Timestamp: d.Timestamp, UserName: d.UserName, Flag: true }, Room: d.Room, Reply: shared.ResponseMD{Status: true}}}
	//send to all users in the room
	for name, user := range s.rooms[d.Room].users {
		//remove user from room state
		remove(name, d.Room)
		if name == d.UserName { //skip if self
			continue
		}
		//notify they are no longer in that room
		user.RecvServer <- force
	}

	//remove room from server state
	delete(s.rooms, d.Room)

	//notify all staff
	msg := formatStaffMsg(d.UserName, "Deleted " + d.Room, d.Timestamp)
	broadcastToStaff(msg)

	d.ErrMsg = "SERVER: Sucessfully deleted " + d.Room
	d.Status = true
}
func (d *DeleteCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// PROMOTE CMD and its execute functions //////////////////////////////
type PromoteDemoteCmd struct {
	*shared.PromoteDemoteCmd
}
func (p *PromoteDemoteCmd) ExecuteServer() {
	s := GetServerState()
	//verify correct usage
	if p.Args != 2 {
		p.Status = false
		p.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	parts := strings.Fields(p.Content)
	p.User = parts[1]
	//verify user can execute this command
	if s.users[p.UserName].Role < RoleOwner {
		p.Status = false
		p.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	}
	//verify specified user exists
	if _, exists := s.users[p.User]; !exists {
		p.Status = false
		p.ErrMsg = "PERMISSION DENIED: User " + p.User + " does not exist" 
		return	
	}
	//check specified user's current role
	if s.users[p.User].Role == RoleAdmin && p.Promote {
		p.Status = false
		p.ErrMsg = "PERMISSION DENIED: User " + p.User + " is already an admin" 
		return	
	} else {
		if s.users[p.User].Role == RoleMember && !p.Promote {
			p.Status = false
			p.ErrMsg = "PERMISSION DENIED: User " + p.User + " is already a member" 
			return	
		} 
	}
	var action, newrole string
	var role Role
	if p.Promote {
		action = "promoted"
		newrole = "to admin"
		role = RoleAdmin
	} else {
		action = "demoted"
		newrole = "to member"
		role = RoleMember

		//if member is currently in a staff only room, kick them out
		if s.users[p.User].CurrentRoom != "" && s.rooms[s.users[p.User].CurrentRoom].permission > RoleMember {
			force := &LeaveCmd{LeaveCmd: &shared.LeaveCmd{ MsgMetadata: shared.MsgMetadata{ Timestamp: p.Timestamp, UserName: p.User, Flag: true }, Room: s.users[p.User].CurrentRoom, Reply: shared.ResponseMD{Status: true}}}
			s.users[p.User].RecvServer <- force
			remove(p.User, s.users[p.User].CurrentRoom)
		}
	}

	//promote member to admin
	s.users[p.User].updateUserState(role)

	//notify staff
	msg := formatStaffMsg(p.UserName, action + " " + p.User + " " + newrole, p.Timestamp)
	broadcastToStaff(msg)
	p.Status = true
	p.ErrMsg = "SERVER: User " + p.User + " was successfully " + action + " " + newrole 
}
func (p *PromoteDemoteCmd) ExecuteClient() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// BROADCAST CMD and its execute functions ////////////////////////////
type BroadcastCmd struct {
	*shared.BroadcastCmd
}
func (b* BroadcastCmd) ExecuteServer() {
	s := GetServerState()
	//verify correct usage
	if b.Args != 2 {
		b.Status = false
		b.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	parts := strings.SplitN(b.Content, " ", 2)
	//take "/broadcast out of content string"
	b.Content = parts[1]
	//verify user can execute this command
	if s.users[b.UserName].Role < RoleAdmin {
		b.Status = false
		b.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	}
	b.Status = true
	//user is at least admin and can broadcast -> send to all ACTIVE users
	for username, user := range s.users {
		//if the user is not active or we are looking at the sender, skip
		if username == b.UserName || !user.Active {
			continue
		}
		//otherwise, send to the user
		user.RecvServer <- b
	}
}
func (b* BroadcastCmd) ExecuteClient() {}
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
		log.Println(user.Role)
		if user.Role >= RoleAdmin && user.Active{
			user.RecvServer <- msg
		}
	}
}

func convToRole(input string) Role {
	if input == "staff" {
		return RoleAdmin
	} else if input == "all" {
		return RoleMember
	} else {
		return RoleBanned //use to indicate the input was incorrect
	}
}