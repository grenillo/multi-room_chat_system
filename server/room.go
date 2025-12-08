package server

import (
	"multi-room_chat_system/shared"
)

//room state
type Room struct {
	//active users
	users map[string]*Member
	//log of messages
	log []shared.Message
	//required room permission
	permission Role
}

//broadcast to all users in a room
func (rm *Room) broadcast(msg *Message, sender string) {
	for username, member := range rm.users {
		//dont send self username, do not send sender username
        if msg.UserName == username || username == sender {
			continue
		}
		//send message to all connected users
		member.RecvServer <- msg
    }
}

//add a user to the room state
func (rm *Room) addUser(user *Member) {
	rm.users[user.Username] = user
}

//remove a user from the room state
func (rm *Room) removeUser(user *Member) {
	delete(rm.users, user.Username)
}