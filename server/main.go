package server

import (
	"net"
	"log"
	"fmt"
)

func StartServer() {
	s := GetServerState() //se
	//setup listener tcp socket on port 5461
	listener, err := net.Listen("tcp", ":5461")
	//if fail to setup socket
	if err != nil {
        log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	fmt.Println("Server listening on :5461")

	// goroutine to handle shutdown signal
	go func() {
		<-s.term
		log.Println("Server shutting down...")
		listener.Close() //unblock accept
	}()

	//continuously listen for new connections
	for {
		select {
		//if server is terminated, stop accepting connections
		case <-s.term:
			log.Println("Server is no longer accepting connections!")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Failed to accept connection:", err)
				continue
			}
			fmt.Println("Accepted connection", conn)
			go handleNewConnection(conn)
		}
	}
}	