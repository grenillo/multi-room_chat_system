package server

import "time"

type Log struct{
	Event string
	Timestamp time.Time
}

//function to format a log event for the server
func logEvent(event string, time time.Time) Log {
	return Log{Event: event, Timestamp: time}
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