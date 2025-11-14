package server

import (
	"multi-room_chat_system/shared"
)

type Room struct {
	//active users
	users map[string]*Member
	//log of messages
	log []shared.Message
	//required room permission
	permission Role
}

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

func (rm *Room) addUser(user *Member) {
	rm.users[user.Username] = user
}

func (rm *Room) removeUser(user *Member) {
	delete(rm.users, user.Username)
}