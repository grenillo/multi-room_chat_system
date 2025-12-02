package server

import (
	"encoding/json"
	"os"
	"time"
	"multi-room_chat_system/shared"
)

//type for persisting user state
type PersistUser struct {
	Username string
	Role Role
}

//type for persisting room state
type PersistRoom struct {
	Name string
	Permission Role
	Log []PersistMessage
}

//type for persisting message state
type PersistMessage struct {
	Username string
	Timestamp time.Time
	Content string
	Image bool
	Flag bool
}

type PersistState struct {
	Users map[string]PersistUser
	Rooms map[string]PersistRoom
	Log []Log
}

func (s *ServerState) SaveToDisk() error {
	//define persistent state
	p := PersistState{Users: make(map[string]PersistUser), Rooms: make(map[string]PersistRoom), Log: make([]Log, 0)}
	//convert current users to the persistent user state
	for name, user := range s.users {
		p.Users[name] = PersistUser{Username: name, Role: user.Role}
	}
	//convert current rooms into the persistent room state
	for name, room := range s.rooms {
		roomInfo := PersistRoom{Name: name, Permission: room.permission, Log: make([]PersistMessage, 0)}
		//loop through the room's current log
		for _, msg := range room.log {
			//convery to persistent message type
			roomInfo.Log = append(roomInfo.Log, PersistMessage{Username: msg.UserName, Timestamp: msg.Timestamp, Content: msg.Content, Image: msg.Image, Flag: msg.Flag})
		}
		//save information to persistent state
		p.Rooms[name] = roomInfo
	}
	//add logger to persistent state
	p.Log = append(p.Log, s.logger...)
	//encode persistent state as JSON
	data, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return err
	}
	//write to file
	return os.WriteFile("serverState.json", data, 0644)
}

func (s *ServerState) LoadFromDisk() error {
	//read from the serverState file
	data, err := os.ReadFile("serverState.json")
	if err != nil {
        return err
    }
	var p PersistState
	//unmarshall json file
    if err := json.Unmarshal(data, &p); err != nil {
        return err
    }
	//rebuild users
	for name, user := range p.Users {
		u := &Member{User: User{Username: name, Role: user.Role, Active: false}}
		//add user back to the server state
		s.users[name] = u
	}
	//rebuild rooms
	for name, room := range p.Rooms {
		r := &Room{users: make(map[string]*Member), log: make([]shared.Message, 0), permission: room.Permission}
		//rebuild room's log
		for _, msg := range room.Log {
			r.log = append(r.log, shared.Message{MsgMetadata: shared.MsgMetadata{UserName: msg.Username, Timestamp: msg.Timestamp, Content: msg.Content, Flag: msg.Flag}, Image: msg.Image})
		}
		//add room back to server state
		s.rooms[name] = r
	}
	//rebuild logger
	s.logger = append(s.logger, p.Log...)
	
	return nil
}