package client

import (
	"fmt"
	"log"
	"multi-room_chat_system/shared"
	//"fyne.io/fyne/v2/internal/widget"
)

/*
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
*/

/////////////////////////////// MESSAGE and its execute functions ///////////////////////////////
type Message struct {
	*shared.Message
}
func (m *Message) ExecuteServer() {}
func (m *Message) ExecuteClient(ui shared.ClientUI)() {
	//check if message was sent
	if !m.Response.Status {
		//fmt.Println(m.Response.ErrMsg)
		ui.Display(m.Response.CurrentRoom, m.Response.ErrMsg)
		return
	}
	log.Println("client received:", m.Message.Content)
	log.Println(m.Response.CurrentRoom)
	ui.Display(m.Response.CurrentRoom, formatMessage(false, m.Message, nil))
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// JOIN CMD and its execute functions ///////////////////////////////
type JoinCmd struct {
	*shared.JoinCmd
}
func (j *JoinCmd) ExecuteServer() {}
func (j *JoinCmd) ExecuteClient(ui shared.ClientUI)() {
	if !j.Reply.Status {
		ui.Display(j.Reply.CurrentRoom, j.Reply.ErrMsg)
		return
	}
	//clear local room history
	ui.ClearRoom(j.Reply.CurrentRoom)
	ui.SelectRoom(j.Reply.CurrentRoom)
	ui.Display(j.Reply.CurrentRoom, "======= JOINED ROOM " + j.Room + " =======")
	//print out entire message history to client
	for _, msg := range j.Reply.Log {
		ui.Display(j.Reply.CurrentRoom, formatMessage(false, &msg, nil))
		//fmt.Println(formatMessage(false, &msg, nil))
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// LEAVE CMD and its execute functions //////////////////////////////
type LeaveCmd struct {
	*shared.LeaveCmd
}

func (l *LeaveCmd) ExecuteServer() {}
func (l *LeaveCmd) ExecuteClient(ui shared.ClientUI)() {
	if !l.Reply.Status {
		ui.Display(l.Reply.CurrentRoom, l.Reply.ErrMsg)
		return
	}
	//clear local room and lobby history
	ui.DeselectRoom()
	ui.ClearRoom(l.Room)
	ui.ClearLobby()
	ui.ShowLobby()
	ui.Display("", "======= LEFT ROOM " + l.Room + " =======")
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// LISTUSERS CMD and its execute functions ////////////////////////////
type ListUsersCmd struct {
	*shared.ListUsersCmd
}

func (lu *ListUsersCmd) ExecuteServer() {}
func (lu *ListUsersCmd) ExecuteClient(ui shared.ClientUI)() {
	if !lu.Reply.Status {
		//fmt.Println(lu.Reply.ErrMsg)
		log.Println("current room:", lu.Reply.CurrentRoom)
		ui.Display(lu.Reply.CurrentRoom, lu.Reply.ErrMsg)
		return
	}
	//if successful print the current room and its users to the client
	//fmt.Println("=======CURRENT USERS IN ROOM ",lu.Reply.Room,"=======")
	ui.Display(lu.Reply.CurrentRoom, "======= CURRENT USERS IN ROOM " + lu.Reply.Room + " =======")
	for _, user := range lu.Reply.Users {
		ui.Display(lu.Reply.CurrentRoom, "\t" + user)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// HELP CMD and its execute functions ///////////////////////////////
type HelpCmd struct {
	*shared.HelpCmd
}
func (h *HelpCmd) ExecuteServer() {}
func (h *HelpCmd) ExecuteClient(ui shared.ClientUI)() {
	if h.Invalid {
		ui.Display(h.Reply.CurrentRoom, h.Reply.ErrMsg)
		return
	}
	ui.Display(h.Reply.CurrentRoom, "======= USER COMMAND USAGE =======")
	//print the commands for this user
	for _, cmd := range h.Reply.Usage {
		ui.Display(h.Reply.CurrentRoom, "\t" + cmd)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////// QUIT CMD and its execute functions ///////////////////////////////
type QuitCmd struct {
	*shared.QuitCmd
}
//user will always be able to quit
func (q *QuitCmd) ExecuteServer() {}
func (q *QuitCmd) ExecuteClient(ui shared.ClientUI)() {
	ui.UserQuit("You have been disconnected from the server")
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// KICK/BAN CMD and its execute functions /////////////////////////////
type KickBanCmd struct {
	*shared.KickBanCmd
}
//user will always be able to quit
func (kb *KickBanCmd) ExecuteServer() {}
func (kb *KickBanCmd) ExecuteClient(ui shared.ClientUI)() {
	if !kb.Status {
		ui.Display(kb.CurrentRoom, kb.ErrMsg)
		return
	}
	if kb.Ban {
		ui.Display(kb.CurrentRoom, "[SERVER] " + kb.User + " was banned successfully")
		//fmt.Println("SERVER:", kb.User ,"was banned successfully")
		return
	}
	ui.Display(kb.CurrentRoom, "[SERVER] " + kb.User + " was kicked successfully")
	//fmt.Println("SERVER:", kb.User,"was kicked successfully")
}

/////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////// CREATE CMD and its execute functions //////////////////////////////
type CreateCmd struct {
	*shared.CreateCmd
}
func (c *CreateCmd) ExecuteServer() {}
func (c *CreateCmd) ExecuteClient(ui shared.ClientUI)() {
	//fmt.Println(c.ErrMsg)
	ui.Display(c.CurrentRoom, c.ErrMsg)
	//if successful, update GUI
	if c.Status {
		ui.AddRoom(c.Room)
	}
}
/////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////// DELETE CMD and its execute functions //////////////////////////////
type DeleteCmd struct {
	*shared.DeleteCmd
}
func (d *DeleteCmd) ExecuteServer() {}
func (d *DeleteCmd) ExecuteClient(ui shared.ClientUI)() {
	if d.InRoom {
		//ClearScreen()
		//fmt.Println("======= LEFT ROOM ", d.Room,"=======")
		ui.ClearRoom(d.Room)
		ui.DeselectRoom()
		ui.Display(d.CurrentRoom, "======= LEFT ROOM " + d.Room + " =======")
	}
	ui.Display(d.CurrentRoom, d.ErrMsg)
	//if successful, update GUI
	if d.Status {
		ui.RemoveRoom(d.Room)
	}
	//fmt.Println(d.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// PROMOTE CMD and its execute functions //////////////////////////////
type PromoteDemoteCmd struct {
	*shared.PromoteDemoteCmd
}
func (p *PromoteDemoteCmd) ExecuteServer() {}
func (p *PromoteDemoteCmd) ExecuteClient(ui shared.ClientUI)() {
	//fmt.Println(p.ErrMsg)
	ui.Display(p.CurrentRoom, p.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// BROADCAST CMD and its execute functions ////////////////////////////
type BroadcastCmd struct {
	*shared.BroadcastCmd
}
func (b* BroadcastCmd) ExecuteServer() {}
func (b* BroadcastCmd) ExecuteClient(ui shared.ClientUI)() {
	//check if message was sent
	if !b.Status {
		ui.Display(b.CurrentRoom, b.ErrMsg)
		//fmt.Println(b.ErrMsg)
		return
	}
	//otherwise, print to our client's local terminal
	ui.Display(b.CurrentRoom, formatMessage(true, nil, b.BroadcastCmd))
}
/////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////// SHUTDOWN CMD and its execute functions ////////////////////////////
type ShutdownCmd struct {
	*shared.ShutdownCmd
}
func (s *ShutdownCmd) ExecuteServer() {}
func (s *ShutdownCmd) ExecuteClient(ui shared.ClientUI)() {
	//fmt.Println(s.ErrMsg)
	ui.Display(s.CurrentRoom, s.ErrMsg)
}
/////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////// LISTROOMS CMD and its execute functions ////////////////////////////
type ListRoomsCmd struct {
	*shared.ListRoomsCmd
}
func (lr *ListRoomsCmd) ExecuteServer() {}
func (lr *ListRoomsCmd) ExecuteClient(ui shared.ClientUI)() {
	ui.Display(lr.CurrentRoom, "===================================================")
	ui.Display(lr.CurrentRoom, lr.ErrMsg)
	ui.Display(lr.CurrentRoom, "===================================================")
}
/////////////////////////////////////////////////////////////////////////////////////////////////

type RoomUpdate struct {
	*shared.RoomUpdate
}
func (ru *RoomUpdate) ExecuteServer() {}
func (ru *RoomUpdate) ExecuteClient(ui shared.ClientUI) {
	//update user interface
	if ru.Create {
		ui.AddRoom(ru.Room)
	} else {
		ui.RemoveRoom(ru.Room)
	}
}

type UserUpdate struct {
	*shared.UserUpdate
}
func (u *UserUpdate) ExecuteServer() {}
func (u *UserUpdate) ExecuteClient(ui shared.ClientUI) {
	//update user interface
	for _, room := range u.Rooms {
		if u.Promote {
			ui.AddRoom(room)
		} else {
			ui.RemoveRoom(room)
		}
	}
}


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