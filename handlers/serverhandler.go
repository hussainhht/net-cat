package handlers

import (
	"fmt"
	"net"
	"os"
	"sync"
)

const defaultPort string = ":8989"

func (msg Message) String() string {
	return fmt.Sprintf("[%s][%s]: %s",
		msg.Timestamp.Format("2006-01-02 15:04:05"),
		msg.Sender.Name,
		msg.Content)
}

var clients []Client
var clientsMutex = &sync.Mutex{}

func RunTCPServer(port string) error {

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close() // we clean the port when done with program

	fmt.Println("Server listening on port" + port)

	// waiting for clients
	for {
		fmt.Println("Waiting for connection")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}

		go HandleClientConnection(conn, &clients, clientsMutex)
	}
}

func GetPort() string {
	Args := os.Args
	if len(Args) > 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := defaultPort
	if len(Args) == 2 {
		port = ":" + Args[1]
	}
	return port
}
