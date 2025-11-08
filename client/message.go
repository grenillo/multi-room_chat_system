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
	fmt.Println(formatMessage(m.Message))
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
		fmt.Println(formatMessage(&msg))
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

func ClearScreen() {
    fmt.Print("\033[2J\033[H\n")
}

func formatMessage(m *shared.Message) string {
	//convert timestamp to string
	time := m.Timestamp.Format("2006-01-02 15:04:05")
	var resp string
	if m.Flag {
		resp = time + "\t" + m.UserName + m.Content
	} else {
		resp = time + "\t" + m.UserName + ":  " + m.Content
	}
	return resp
}