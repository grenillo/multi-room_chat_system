package server

import(
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

func (rm *Room) broadcast(msg *Message) {
	for username, member := range rm.users {
		//dont send self username
        if msg.UserName == username {
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