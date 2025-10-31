package server

import (
	"log"
	"multi-room_chat_system/shared"
	"sync"
	"time"
)

//internal state for the server
type ServerState struct {
	//map usernames to Members
	users map[string]*Member
	//map roomName to chatRoom
	rooms map[string]*Room
	//logger
	//logger *Logger
	//server configuration
	//config *Config
	//server dispatcher
	//dispatcher *Dispatcher

	//channels to receive/respond user joins 
	recvUser chan ServerJoinRequest
	joinResp chan *ServerJoinResponse

	recvRoom chan *Room
	recvInput chan *shared.MsgMetadata
	ackInput chan *shared.ExecutableMessage

	//recvLogger chan *Logger
	//recvDispatcher chan *Dispatcher

	//channel to handle joining rooms
	//recvJoinReq chan JoinRoomReq
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
		users: map[string]*Member{},
		rooms: map[string]*Room{},
		//logger: &Logger{},
		//config: &Config{},
		//dispatcher: &Dispatcher{},
		//channels for joining users
		recvUser: make(chan ServerJoinRequest),
		joinResp: make(chan *ServerJoinResponse),

		recvRoom: make(chan *Room),
		//channels for message input
		recvInput: make(chan *shared.MsgMetadata),
		ackInput: make(chan *shared.ExecutableMessage),
		//recvLogger: make(chan *Logger),
		//recvDispatcher: make(chan *Dispatcher),
	}
	//create initial rooms
	instance.createRoom("#general")
	instance.createRoom("#staff")
	//start goroutine to run server
	go instance.run()	
	
}

func(s *ServerState) run() {
	for{
		select {
		//server management of users
		case userState := <-s.recvUser:
			//response variable
			var resp ServerJoinResponse
			//check if user exists
			username := string(userState)
			if _, exists := s.users[username]; !exists {
				//if dne create a new user of type member
				newUser := UserFactory(username, RoleMember)
				//add new user to the server state
				s.users[username] = newUser
				resp = ServerJoinResponse{
					Status: true,
					Message: "Welcome to the server!\n",
					Role: newUser,
				}
			//if user already exists
			} else {
				if s.users[username].Role != RoleBanned {
					resp = ServerJoinResponse{
						Status: true,
						Message: "Welcome back to the server!\n",
						Role: s.users[username],
					}
				} else {
					resp = ServerJoinResponse{
						Status: false,
						Message: "You are banned!\n",
						Role: s.users[username],
					}
				}
			}
			//send response
			s.joinResp <- &resp
		//server receives raw input from a client
		case input := <-s.recvInput:
			//add timestamp to metadata
			input.Timestamp = time.Now()
			//get user's current room

			//call message factory
			msg := MessageFactory(*input, s)
			//execute the msg
			msg.ExecuteServer()
			//ack the RPC
			s.ackInput <- &msg

		default:
		}
	}

}

//helper function to create a room
func (s *ServerState) createRoom(roomName string) {
	if _, exists := s.rooms[roomName]; exists {
		log.Println("Error: room already exists!")
		return
	}
	//initialize the room's state
	newRoom := Room{
		users: make(map[string]*Member),
		log: make([]shared.Message, 0),
	}
	//add new room to the server's state
	s.rooms[roomName] = &newRoom
}

//helper function to remove a room
func (s *ServerState) deleteRoom(roomName string) {
	if _, exists := s.rooms[roomName]; !exists {
		log.Println("Error: room does not exist!")
		return
	}
	//remove it from the server's state
	delete(s.rooms, roomName)
}


//join server RPC stub
func (s *ServerState) JoinServer(username string, reply *ServerJoinResponse) error {
	//create join request
    req := ServerJoinRequest(username)
    // Send to the server's state goroutine
    s.recvUser <- req
    // Wait for the server to respond
	temp := <-s.joinResp
	*reply = *temp
    return nil
}

//receive input RPC stub
func (s *ServerState) RecvMessage(input *shared.MsgMetadata, reply *shared.ExecutableMessage) error {
	//send metadata to the server
	s.recvInput <- input
	//wait for ack
	resp := <- s.ackInput
	*reply = *resp
	return nil
}