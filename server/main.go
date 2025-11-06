package server

import (
	"net"
	"log"
	"fmt"
)

func StartServer() {
	GetServerState() //se
	//setup listener tcp socket on port 5461
	listener, err := net.Listen("tcp", ":5461")
	//if fail to setup socket
	if err != nil {
        log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	fmt.Println("Server listening on :5461")
	//continuously listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		fmt.Println("Accepted connection", conn)
		handleNewConnection(conn)
	}
}	