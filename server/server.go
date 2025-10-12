package server

import (
	"log"
	"sync"
	"multi-room_chat_system/common"
)

//internal state for the server
type ServerState struct {
	//map usernames to Members
	users map[string]*User
	//map roomName to chatRoom
	rooms map[string]*Room
	//logger
	logger *Logger
	//server configuration
	config *Config
	//server dispatcher
	dispatcher *Dispatcher
	//channels to receive information
	recvUser chan *common.JoinRequest
	recvRoom chan *Room
	recvLogger chan *Logger
	recvDispatcher chan *Dispatcher
}

//singleton instance for the server
var(
	instance *ServerState
	once sync.Once
)

//get the singleton instance for the server
func GetServerState() *ServerState {
	once.Do(initServer)
	return instance
}

//function to initialize the server state, and starts up the goroutine that manages server state
func initServer() {
	log.Println("Starting server")
	instance = &ServerState{
		users: map[string]*User{},
		rooms: map[string]*Room{},
		logger: &Logger{},
		config: &Config{},
		dispatcher: &Dispatcher{},
		recvUser: make(chan *common.JoinRequest),
		recvRoom: make(chan *Room),
		recvLogger: make(chan *Logger),
		recvDispatcher: make(chan *Dispatcher),
	}
	//start goroutine to run server
	go instance.run()	
	
}

func(s *ServerState) run() {
	for{
		select {
		//server management of users
		case userState := <-s.recvUser:
			resp := common.JoinResponse{}
			//first check to see if this user already exists
			if _, exists := s.users[userState.UserName]; !exists {
				//user does not exist, create a new role as a member for the user
				member := UserFactory(userState.UserName, common.RoleMember)
				resp.Role = common.RoleMember
				resp.RoleObj = member
				resp.Message = "Welcome to the server!\n"
				resp.Status = true
				//TODO add initalization of the general chatroom, with a RoomInfo type

				//add user to the map of all users on the server
				//s.users[userState.UserName] = member

			//if the username does exist, return the role to the user
			} else {
				role := UserFactory(userState.UserName, s.users[userState.UserName].Role)
				resp.Role = s.users[userState.UserName].Role
				resp.RoleObj = role
				if role != common.RoleBanned {
					resp.Message = "Welcome back to the server!\n"
					resp.Status = true
				} else {
					resp.Message = "You are banned from the server!\n"
					resp.Status = false
				}
			}
			//send response
			userState.Response <- &resp
			

		//server management of rooms 
		//case roomState := <-s.recvRoom:
			//temp
		//server management of logging
		//case log := <-s.recvLogger:
			//temp
		//server management of dispatching
		//case dispatch := <-s.recvDispatcher:
			//temp
		default:
		}
	}

}


//join server RPC stub
func (s *ServerState) JoinServer(username string, reply *common.JoinResponse) error {
    replyCh := make(chan *common.JoinResponse)
	//create join request
    req := &common.JoinRequest{UserName: username, Response: replyCh}

    // Send to the server's state goroutine
    s.recvUser <- req
    // Wait for the server to respond
	temp := <-replyCh
	*reply = *temp

    return nil
}