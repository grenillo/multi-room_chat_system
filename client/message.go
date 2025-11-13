package client

import (
	"fmt"
	"multi-room_chat_system/shared"
)

/////////////////////////////// MESSAGE and its execute functions ///////////////////////////////
type Message struct {
	*shared.Message
}
func (m *Message) ExecuteServer() {}
func (m *Message) ExecuteClient() {
	//check if message was sent
	if !m.Response.Status {
		fmt.Println(m.Response.ErrMsg)
		return
	}
	//otherwise, print to our client's local terminal
	fmt.Println(formatMessage(false, m.Message, nil))
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// JOIN CMD and its execute functions ///////////////////////////////
type JoinCmd struct {
	*shared.JoinCmd
}
func (j *JoinCmd) ExecuteServer() {}
func (j *JoinCmd) ExecuteClient() {
	if !j.Reply.Status {
		fmt.Println(j.Reply.ErrMsg)
		return
	}
	ClearScreen()
	fmt.Println("=======JOINED ROOM ", j.Room, "=======")
	//print out entire message history to client
	for _, msg := range j.Reply.Log {
		fmt.Println(formatMessage(false, &msg, nil))
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// LEAVE CMD and its execute functions //////////////////////////////
type LeaveCmd struct {
	*shared.LeaveCmd
}

func (l *LeaveCmd) ExecuteServer() {}
func (l *LeaveCmd) ExecuteClient() {
	if !l.Reply.Status {
		fmt.Println(l.Reply.ErrMsg)
		return
	}
	ClearScreen()
	fmt.Println("=======LEFT ROOM ", l.Room,"=======")	
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// LISTUSERS CMD and its execute functions ////////////////////////////
type ListUsersCmd struct {
	*shared.ListUsersCmd
}

func (lu *ListUsersCmd) ExecuteServer() {}
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
	*shared.HelpCmd
}
func (h *HelpCmd) ExecuteServer() {}
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

////////////////////////////// QUIT CMD and its execute functions ///////////////////////////////
type QuitCmd struct {
	*shared.QuitCmd
}
//user will always be able to quit
func (q *QuitCmd) ExecuteServer() {}
func (q *QuitCmd) ExecuteClient() {
	ClearScreen()
	fmt.Println("====================== GOODBYE ======================")
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// KICK/BAN CMD and its execute functions /////////////////////////////
type KickBanCmd struct {
	*shared.KickBanCmd
}
//user will always be able to quit
func (kb *KickBanCmd) ExecuteServer() {}
func (kb *KickBanCmd) ExecuteClient() {
	if !kb.Status {
		fmt.Println(kb.ErrMsg)
		return
	}
	if kb.Ban {
		fmt.Println("SERVER:", kb.User ,"was banned successfully")
		return
	}
	fmt.Println("SERVER:", kb.User,"was kicked successfully")
}

/////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////// CREATE CMD and its execute functions //////////////////////////////
type CreateCmd struct {
	*shared.CreateCmd
}
func (c *CreateCmd) ExecuteServer() {}
func (c *CreateCmd) ExecuteClient() {
	fmt.Println(c.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// DELETE CMD and its execute functions //////////////////////////////
type DeleteCmd struct {
	*shared.DeleteCmd
}
func (d *DeleteCmd) ExecuteServer() {}
func (d *DeleteCmd) ExecuteClient() {
	if d.InRoom {
		ClearScreen()
		fmt.Println("======= LEFT ROOM ", d.Room,"=======")	
	}
	fmt.Println(d.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// PROMOTE CMD and its execute functions //////////////////////////////
type PromoteDemoteCmd struct {
	*shared.PromoteDemoteCmd
}
func (p *PromoteDemoteCmd) ExecuteServer() {}
func (p *PromoteDemoteCmd) ExecuteClient() {
	fmt.Println(p.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// BROADCAST CMD and its execute functions ////////////////////////////
type BroadcastCmd struct {
	*shared.BroadcastCmd
}
func (b* BroadcastCmd) ExecuteServer() {}
func (b* BroadcastCmd) ExecuteClient() {
	//check if message was sent
	if !b.Status {
		fmt.Println(b.ErrMsg)
		return
	}
	//otherwise, print to our client's local terminal
	fmt.Println(formatMessage(true, nil, b.BroadcastCmd))
}
/////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////// SHUTDOWN CMD and its execute functions ////////////////////////////
type ShutdownCmd struct {
	*shared.ShutdownCmd
}
func (s *ShutdownCmd) ExecuteServer() {}
func (s *ShutdownCmd) ExecuteClient() {
	fmt.Println(s.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// LISTROOMS CMD and its execute functions ////////////////////////////
type ListRoomsCmd struct {
	*shared.ListRoomsCmd
}
func (lr *ListRoomsCmd) ExecuteServer() {}
func (lr *ListRoomsCmd) ExecuteClient() {
	fmt.Println("===================================================")
	fmt.Println(lr.ErrMsg)
	fmt.Println("===================================================")
}
/////////////////////////////////////////////////////////////////////////////////////////////////


func ClearScreen() {
    fmt.Print("\033[2J\033[H\n")
}

func formatMessage(broadcast bool, m *shared.Message, b *shared.BroadcastCmd) string {
	var resp string
	if broadcast {
		//convert timestamp to string
		time := b.Timestamp.Format("2006-01-02 15:04:05")
		if b.Flag {
			resp = time + "\t" + b.UserName + b.Content
		} else {
			resp = time + "\t" + b.UserName + ":  " + b.Content
		}
	} else {
		//convert timestamp to string
		time := m.Timestamp.Format("2006-01-02 15:04:05")
		if m.Flag {
			resp = time + "\t" + m.UserName + m.Content
		} else {
			resp = time + "\t" + m.UserName + ":  " + m.Content
		}
	}
	return resp
}