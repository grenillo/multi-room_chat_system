package server

import (
	"fmt"
	"log"
	"multi-room_chat_system/shared"
	"net/http"
	"os"
	"sync"
	"time"
)

//internal state for the server
type ServerState struct {
	shutdownReq bool
	//map usernames to Members
	users map[string]*Member
	//map roomName to chatRoom
	rooms map[string]*Room
	//file server for image support
	fileServer *http.Server
	//logger for server
	logger []Log
	//channels to receive/respond user joins 
	recvUser chan ServerJoinRequest
	joinResp chan *ServerJoinResponse
	
	recvInput chan *shared.MsgMetadata
	ackInput chan *shared.ExecutableMessage
	term chan struct{}
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
	shared.Init()
	log.Println("Starting server")
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Could not create uploads dir:", err)
	}
	instance = &ServerState{
		shutdownReq: false,
		users: map[string]*Member{},
		rooms: map[string]*Room{},
		//channels for joining users
		recvUser: make(chan ServerJoinRequest),
		joinResp: make(chan *ServerJoinResponse),
		//channels for message input
		recvInput: make(chan *shared.MsgMetadata),
		ackInput: make(chan *shared.ExecutableMessage),
		term: make(chan struct{}),
		logger: make([]Log, 0),
	}
	instance.fileServer = startFileServer()
	instance.LoadFromDisk()
	//start goroutine to run server
	go instance.run()	
	
}

func(s *ServerState) run() {
	//add admin
	//instance.users["owner"] = UserFactory("owner", RoleOwner)
	//instance.users["admin"] = UserFactory("admin", RoleAdmin)
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
				newUser.Active = true
				s.users[username] = newUser
				rooms := getJoinableRooms(s.users[username])
				resp = ServerJoinResponse{
					Status: true,
					Message: "Welcome to the server!\n" + rooms,
					Role: newUser,
				}
			//if user already exists
			} else {
				//first check to see if the user is currently logged in
				if s.users[username].Active {
					resp = ServerJoinResponse{
						Status: false,
						Message: "PERMISSION DENIED: " + username + " is currently logged in!\n>",
						Role: s.users[username],
					}
				} else { //user not logged in
					if s.users[username].Role != RoleBanned {
						//create new object
						user := UserFactory(username, s.users[username].Role)
						//add user to the server state for updated channels
						user.Active = true
						s.users[username] = user
						rooms := getJoinableRooms(s.users[username])
						resp = ServerJoinResponse{
							Status: true,
							Message: "Welcome back to the server!\n" + rooms,
							Role: s.users[username],
						}
					} else {
						resp = ServerJoinResponse{
							Status: false,
							Message: "PERMISSION DENIED: You are banned!\n>",
							Role: s.users[username],
						}
					}
				}
			}
			//if status is true log that the user joined
			if resp.Status {
				s.logger = append(s.logger, logEvent(username + " joined the server", time.Now()))
			}
			//send response
			s.joinResp <- &resp
			//if admin send log
			if resp.Status && resp.Role.Role > RoleMember {
				resp.Role.RecvServer <- &GetLog{&shared.GetLog{Log: s.formatLog()}}
			}
		//server receives raw input from a client
		case input := <-s.recvInput:
			//add timestamp to metadata
			input.Timestamp = time.Now()
			log.Println("server received:", input.Content)
			//call message factory
			msg := MessageFactory(*input, s)
			log.Println("server generated factory object for:", input.Content)
			//execute the msg
			msg.ExecuteServer()
			//ack the RPC
			s.ackInput <- &msg
			if s.shutdownReq {
				//write back the server's current state
				log.Println("SERVER: attempting to save state")
				s.SaveToDisk()
				log.Println("SERVER: state save successful")
				close(s.term)
			}
		case <-s.term:
			log.Println("SERVER: terminated, returning")
			s.fileServer.Close()
			return
		}
	}

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

func getJoinableRooms(user *Member) string {
	temp := "Available rooms:"
	for _, room := range user.AvailableRooms {
		temp += " " + room
	}
	temp += ">"
	return temp 
}

func startFileServer() *http.Server {
    mux := http.NewServeMux()

    // Serve static files from ./uploads
    fs := http.FileServer(http.Dir("./uploads"))
    mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	mux.HandleFunc("/upload", uploadHandler)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    go func() {
        fmt.Println("HTTP file server running at http://localhost:8080/uploads/")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Println("File server error:", err)
        }
    }()

    return srv
}
