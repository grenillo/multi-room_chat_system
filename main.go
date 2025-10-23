package main

import (
	"multi-room_chat_system/server"
	"net"
	"log"
	"fmt"
)

func main() {
	//get the server state
	s := server.GetServerState()
	//setup tcp socket
	ln, err := net.Listen("tcp", ":8080")
	//if fail to setup socket
	if err != nil {
        log.Fatalf("Error starting server: %v", err)
	}
	defer ln.Close()
	fmt.Println("Server listening on :8080")

	//start listening loop from server/rooms
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go server.HandleConnection(conn, s)
	}
}	