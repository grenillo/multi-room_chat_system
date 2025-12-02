package server

import (
	"multi-room_chat_system/shared"
	"time"
)

type Log struct{
	Event string
	Timestamp time.Time
}

//function to format a log event for the server
func logEvent(event string, time time.Time, sender string) Log {
	log := Log{Event: event, Timestamp: time}
	broadcastStaffLobby(sender, log)
	return log
}

func (s *ServerState) formatLog() []string {
	log := []string{}
	s = GetServerState()
	//loop through logger
	for _, l := range s.logger {
		var temp string
		time := l.Timestamp.Format("2006-01-02 15:04:05")
		temp = time + "\t\t" + l.Event
		log = append(log, temp)
	}
	return log
}

func broadcastStaffLobby(sender string, log Log) {
	s := GetServerState()
	for name, user := range s.users {
		//if user is not an admin or we are looking at the sender, do not send update
		if !user.Active || user.Role < RoleAdmin || name == sender {
			continue
		} else { //at least an admin and not self
			if user.CurrentRoom == "" {
				var temp string
				time := log.Timestamp.Format("2006-01-02 15:04:05")
				temp = time + "\t\t" + log.Event
				user.RecvServer <- &UpdateLobby{&shared.UpdateLobby{Update: temp}}
			}
		}

	}
}