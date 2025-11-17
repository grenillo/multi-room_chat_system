package server

import (
	"log"
	"multi-room_chat_system/shared"
	//"runtime/trace"
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
	case "/unban":
		return &UnBanCmd{UnBanCmd: &shared.UnBanCmd{MsgMetadata: input}}
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
	case "/shutdown":
		return &ShutdownCmd{ShutdownCmd: &shared.ShutdownCmd{MsgMetadata: input}}
	case "/listrooms":
		return &ListRoomsCmd{ListRoomsCmd: &shared.ListRoomsCmd{MsgMetadata: input}}
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
		resp = shared.ResponseMD{Status: false, ErrMsg: "PERMISSION DENIED: User is not currently in a room", CurrentRoom: ""}
		m.Response = resp
		return
	}
	//user in room, log message
	s.rooms[s.users[m.UserName].CurrentRoom].log = append(s.rooms[s.users[m.UserName].CurrentRoom].log, *m.Message)
	//broadcast to all other users
	resp = shared.ResponseMD{Status: true, CurrentRoom: s.users[m.UserName].CurrentRoom}
	m.Response = resp
	log.Println("Message room:", resp.CurrentRoom)
	s.rooms[s.users[m.UserName].CurrentRoom].broadcast(m, "")
}

func (m *Message) ExecuteClient(ui shared.ClientUI) {}
/////////////////////////////////////////////////////////////////////////////////////////////////


////////////////////////////// JOIN CMD and its execute functions //////////////////////////////
type JoinCmd struct {
	*shared.JoinCmd
}

func (j *JoinCmd) ExecuteServer() {
	s := GetServerState()
	//check that the cmd was entered properly
	if j.Args != 2 {
		j.Reply.CurrentRoom = s.users[j.UserName].CurrentRoom
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
		j.Reply.CurrentRoom = s.users[j.UserName].CurrentRoom
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: Room does not exist"
		return
	}
	//check if the user is already in this room
	log.Println("currRoom:", s.users[j.UserName].CurrentRoom)
	log.Println("jRoom:", j.Room)
	if s.users[j.UserName].CurrentRoom == j.Room {
		j.Reply.CurrentRoom = s.users[j.UserName].CurrentRoom
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: User already in specified room"
		return
	}
	//check if the user has permission to join this room
	//banned < member < admin < owner, so if user's role is less than the required permission
	if s.users[j.UserName].Role < s.rooms[j.Room].permission {
		j.Reply.CurrentRoom = s.users[j.UserName].CurrentRoom
		j.Reply.Status = false
		j.Reply.ErrMsg = "PERMISSION DENIED: User role does not have access to room"
		return
	}
	j.Reply.CurrentRoom = s.users[j.UserName].CurrentRoom
	//check if user is in a different room, if they are, remove them from that room's state before proceeding
	if s.users[j.UserName].CurrentRoom != "" && j.Room != s.users[j.UserName].CurrentRoom {
		//broadcast user leaving to other members in that room
		broadcast(j.UserName, "left", j.Timestamp, s.users[j.UserName].CurrentRoom, "")
		s.rooms[s.users[j.UserName].CurrentRoom].removeUser(s.users[j.UserName])
	}
	//add user to room
	s.rooms[j.Room].addUser(s.users[j.UserName])
	j.Reply.Status = true
	//update user's room
	s.users[j.UserName].CurrentRoom = j.Room
	j.Reply.CurrentRoom = j.Room

	//broadcast and log user joining to all others currently in the room
	broadcast(j.UserName, "joined", j.Timestamp, j.Room, "")

	//store the room's current state of messages in the response
	j.Reply.Log = s.rooms[j.Room].log

}
func (j *JoinCmd) ExecuteClient(ui shared.ClientUI) {}
/////////////////////////////////////////////////////////////////////////////////////////////////


////////////////////////////// LEAVE CMD and its execute functions //////////////////////////////
type LeaveCmd struct {
	*shared.LeaveCmd
}

func (l *LeaveCmd) ExecuteServer() {
	s := GetServerState()
	l.Reply.CurrentRoom = s.users[l.UserName].CurrentRoom
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
	l.Reply.CurrentRoom = l.Room
	//broadcast and log user leaving to all others currently in the room
	broadcast(l.UserName, "left", l.Timestamp, l.Room, "")


	//update user state
	l.Reply.Status = true
}

func (l *LeaveCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// LISTUSERS CMD and its execute functions ////////////////////////////
type ListUsersCmd struct {
	*shared.ListUsersCmd
}

func (lu *ListUsersCmd) ExecuteServer() {
	s := GetServerState()
	lu.Reply.CurrentRoom = s.users[lu.UserName].CurrentRoom
	//check that the cmd was entered properly
	if lu.Args != 1 {
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

func (lu *ListUsersCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// HELP CMD and its execute functions ///////////////////////////////
type HelpCmd struct {
	*shared.HelpCmd
}

func (h *HelpCmd) ExecuteServer() {
	s := GetServerState()
	h.Reply.CurrentRoom = s.users[h.UserName].CurrentRoom
	if h.Invalid {
		h.Reply.Status = false
		h.Reply.ErrMsg = "PERMISSION DENIED: Invalid command, enter /help for more information"
	}
	//get the usage for this user's role
	h.Reply.Status = true
	h.Reply.Usage = getUsage(s.users[h.UserName].Role)
}
func (h *HelpCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


////////////////////////////// QUIT CMD and its execute functions ///////////////////////////////
type QuitCmd struct {
	*shared.QuitCmd
}
//user will always be able to quit
func (q *QuitCmd) ExecuteServer() {
	s := GetServerState()
	q.CurrentRoom = ""
	//check if user is in a room
	if s.users[q.UserName].CurrentRoom != "" {
		room := s.users[q.UserName].CurrentRoom
		remove(q.UserName, room)
		broadcast(q.UserName, "left", q.Timestamp, room, "")
	}
	//set user status to false
	s.users[q.UserName].Active = false
	//close this connectionHandler once response is sent
	safeClose(s.users[q.UserName].Term)
}
func (q *QuitCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// KICK/BAN CMD and its execute functions /////////////////////////////
type KickBanCmd struct {
	*shared.KickBanCmd
}
func (kb *KickBanCmd) ExecuteServer() {
	s := GetServerState()
	kb.CurrentRoom = s.users[kb.UserName].CurrentRoom
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
			var self *Message
			if s.users[kb.User].CurrentRoom != "" {
				room := s.users[kb.User].CurrentRoom
				remove(kb.User, room)
				self = broadcast(kb.User, "left", kb.Timestamp, room, kb.UserName)
			}
			update := &KickBanCmd{KickBanCmd: &shared.KickBanCmd{Sender: false}}
			update.Status = true
			var msg *Message
			//if ban
			if kb.Ban {
				s.users[kb.User].Role = RoleBanned
				//broadcast ban to staff
				msg = formatStaffMsg(kb.UserName, "banned user: " + kb.User, kb.Timestamp)
				//update = formatStaffMsg("You", "have been banned!", kb.Timestamp)
				update.ErrMsg = "You have been banned!"
			} else {
				msg = formatStaffMsg(kb.UserName, "kicked user: " + kb.User, kb.Timestamp)
				update.ErrMsg = "You have been kicked!"
			}
			//if sender is in the same room as the specified user
			if s.users[kb.UserName].CurrentRoom != "" && s.users[kb.UserName].CurrentRoom ==  s.users[kb.User].CurrentRoom {
				kb.Msg = *self.Message
				kb.InRoom = true
			}
			broadcastToStaff(msg)
			if s.users[kb.User].Active {
				//update user, update its active state, then close its term channel
				s.users[kb.User].RecvServer <- update
				s.users[kb.User].Active = false
				safeClose(s.users[kb.User].Term)
			}
			kb.Sender = true
		}
	}
}
func (kb *KickBanCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// UNBAN CMD and its execute functions ///////////////////////////////
type UnBanCmd struct {
	*shared.UnBanCmd
}
func (u *UnBanCmd) ExecuteServer() {
	s := GetServerState()
	u.CurrentRoom = s.users[u.UserName].CurrentRoom
	//check that the cmd was entered properly
	if u.Args != 2 {
		u.Status = false
		u.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//set user after verifying it exists
	parts := strings.Fields(u.Content)
	u.User = parts[1]
	//first check user's role to see if they can execute
	if (s.users[u.UserName].Role < RoleAdmin) {
		u.Status = false
		u.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	} else { //is an admin or owner
		//check to see if the specified user exists
		if _, exists := s.users[u.User]; !exists {
			u.Status = false
			u.ErrMsg = "PERMISSION DENIED: User: " + u.User + " does not exist on this server"
			return
		} else { //user exists in the banned state
			//update the user's role
			s.users[u.User].Role = RoleMember
			//broadcast to all staff
			msg := formatStaffMsg(u.UserName, "unbanned user: " + u.User, u.Timestamp)
			broadcastToStaff(msg)
			u.ErrMsg = "[SERVER] " + u.User + " successfully unbanned"
		}
	}
}
func (u *UnBanCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////// CREATE CMD and its execute functions //////////////////////////////
type CreateCmd struct {
	*shared.CreateCmd
}
func (c *CreateCmd) ExecuteServer() {
	s := GetServerState()
	c.CurrentRoom = s.users[c.UserName].CurrentRoom
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
	//create live update object
	rmUpdate := &RoomUpdate{RoomUpdate: &shared.RoomUpdate{Create: true, Room: c.Room}}
	//update user states
	for name, user := range s.users {
		//only update if they are at least the correct role
		if user.Role >= newRoom.permission {
			user.AvailableRooms = append(user.AvailableRooms, c.Room)
			//if user is not active or self, do not broadcast live update
			if !user.Active || name == c.UserName {
				continue
			}
			//send update to user
			user.RecvServer <- rmUpdate
			
		}
	}
	broadcastToStaff(formatStaffMsg(c.UserName, "created room " + c.Room, c.Timestamp))
	c.Status = true
	c.ErrMsg = "SERVER: room " + c.Room + " was successfully created"
}
func (c *CreateCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// DELETE CMD and its execute functions //////////////////////////////
type DeleteCmd struct {
	*shared.DeleteCmd
}
func (d *DeleteCmd) ExecuteServer() {
	s := GetServerState()
	d.CurrentRoom = s.users[d.UserName].CurrentRoom
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
	//create live update object
	rmUpdate := &RoomUpdate{RoomUpdate: &shared.RoomUpdate{Create: false, Room: d.Room}}
	//send to all users in the room
	for name, user := range s.rooms[d.Room].users {
		//remove user from room state
		remove(name, d.Room)
		if name == d.UserName { //skip if self
			continue
		}
		//notify they are no longer in that room
		user.RecvServer <- force
		//update user GUI
		user.RecvServer <- rmUpdate
	}

	//remove room from server state
	delete(s.rooms, d.Room)

	//notify all staff
	broadcastToStaff(formatStaffMsg(d.UserName, "deleted room " + d.Room, d.Timestamp))

	d.ErrMsg = "SERVER: Sucessfully deleted " + d.Room
	d.Status = true
}
func (d *DeleteCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// PROMOTE CMD and its execute functions //////////////////////////////
type PromoteDemoteCmd struct {
	*shared.PromoteDemoteCmd
}
func (p *PromoteDemoteCmd) ExecuteServer() {
	s := GetServerState()
	p.CurrentRoom = s.users[p.UserName].CurrentRoom
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
	update := &UserUpdate{UserUpdate: &shared.UserUpdate{Rooms: make([]string, 0)}}
	if p.Promote {
		action = "promoted"
		newrole = "to admin"
		role = RoleAdmin
		update.Promote = true
	} else {
		action = "demoted"
		newrole = "to member"
		role = RoleMember
		update.Promote = false
		//if member is currently in a staff only room, kick them out
		if s.users[p.User].CurrentRoom != "" && s.rooms[s.users[p.User].CurrentRoom].permission > RoleMember {
			force := &LeaveCmd{LeaveCmd: &shared.LeaveCmd{ MsgMetadata: shared.MsgMetadata{ Timestamp: p.Timestamp, UserName: p.User, Flag: true }, Room: s.users[p.User].CurrentRoom, Reply: shared.ResponseMD{Status: true}}}
			s.users[p.User].RecvServer <- force
			remove(p.User, s.users[p.User].CurrentRoom)
		}
	}
	//promote member to admin
	s.users[p.User].updateUserState(role, update)

	//update user's GUI state
	s.users[p.User].RecvServer <- update

	//notify staff
	msg := formatStaffMsg(p.UserName, action + " " + p.User + " " + newrole, p.Timestamp)
	broadcastToStaff(msg)
	p.Status = true
	p.ErrMsg = "SERVER: User " + p.User + " was successfully " + action + " " + newrole 
}
func (p *PromoteDemoteCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// BROADCAST CMD and its execute functions ////////////////////////////
type BroadcastCmd struct {
	*shared.BroadcastCmd
}
func (b* BroadcastCmd) ExecuteServer() {
	s := GetServerState()
	b.CurrentRoom = s.users[b.UserName].CurrentRoom
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
		bc := *b.BroadcastCmd // copy the underlying struct
		cmd := &BroadcastCmd{BroadcastCmd: &bc}
		cmd.CurrentRoom = user.CurrentRoom
		//otherwise, send to the user
		user.RecvServer <- cmd
	}
	b.CurrentRoom = s.users[b.UserName].CurrentRoom
}
func (b* BroadcastCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// SHUTDOWN CMD and its execute functions ////////////////////////////
type ShutdownCmd struct {
	*shared.ShutdownCmd
}
func (sh *ShutdownCmd) ExecuteServer() {
	s := GetServerState()
	sh.CurrentRoom = s.users[sh.UserName].CurrentRoom
	//verify correct usage
	if sh.Args != 1 {
		sh.Status = false
		sh.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	//verify user can execute this command
	if s.users[sh.UserName].Role < RoleOwner {
		sh.Status = false
		sh.ErrMsg = "PERMISSION DENIED: You do not have permission to execute this command"
		return
	}
	sh.Status = true
	sh.Sender = true
	sh.ErrMsg = "SERVER: Shutdown was successful"
	s.shutdownReq = true
	for name, user := range s.users {
		if !user.Active || name == sh.UserName {
			continue
		}
		//send final shutdown to client so their GUI shuts down
		shutdown := &ShutdownCmd{ShutdownCmd: &shared.ShutdownCmd{Sender: false}}
		shutdown.Status = true
		user.RecvServer <- shutdown
	}
}
func (sh *ShutdownCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////


//////////////////////////// LISTROOMS CMD and its execute functions ////////////////////////////
type ListRoomsCmd struct {
	*shared.ListRoomsCmd
}
func (lr *ListRoomsCmd) ExecuteServer() {
	s := GetServerState()
	lr.CurrentRoom = s.users[lr.UserName].CurrentRoom
	//check that the cmd was entered properly
	if lr.Args != 1 {
		lr.Status = false
		lr.ErrMsg = "PERMISSION DENIED: Incorrect usage, enter /help for more information"
		return
	}
	rooms := getRooms(s.users[lr.UserName].Role)
	resp := "Available rooms:\n"
	for i, r := range rooms {
		if i == len(rooms) - 1 {
			resp += "\t" + r
			continue
		}
		resp += "\t" + r + "\n"
	}
	lr.ErrMsg = resp
	lr.Status = true
}
func (lr *ListRoomsCmd) ExecuteClient(ui shared.ClientUI)() {}
/////////////////////////////////////////////////////////////////////////////////////////////////

//stubs for updating a room upon creation/deletion -> mainly used by GUI and create/delete cmds
type RoomUpdate struct {
	*shared.RoomUpdate
}
func (ru *RoomUpdate) ExecuteServer() {}
func (ru *RoomUpdate) ExecuteClient(ui shared.ClientUI) {}

//stubs for updating a user's state upon promotion/demotion -> used internally not for cmds
type UserUpdate struct {
	*shared.UserUpdate
}
func (u *UserUpdate) ExecuteServer() {}
func (u *UserUpdate) ExecuteClient(ui shared.ClientUI) {}

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


func broadcast(username string, action string, timestamp time.Time, room string, sender string) *Message{
	s := GetServerState()
	//add user action to the room's log
	m := shared.Message{
		MsgMetadata: shared.MsgMetadata{
			Timestamp: timestamp,
			UserName:  username,
			Flag:      true,
			Content:   " " + action + " " + room,
		},
		Response: shared.ResponseMD{Status: true, CurrentRoom: room},
	}
	M := &Message{Message: &m}
	s.rooms[room].log = append(s.rooms[room].log, m)

	//broadcast user action to all other users in the room
	log.Println("broadcasting message to room", room)
	s.rooms[room].broadcast(M, sender)
	return M
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
			m := *msg.Message
			message := &Message{Message: &m}
			message.Response.CurrentRoom = user.CurrentRoom
			user.RecvServer <- message
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