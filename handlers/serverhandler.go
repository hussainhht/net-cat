package handlers

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const defaultPort string = ":8989"

type Client struct {
	Name       string
	Connection net.Conn
	Room       Room
}

type Room struct {
	Name    string
	Members []Client
	History []Message
}

type Message struct {
	Timestamp time.Time
	Sender    Client
	Content   string
}

func (msg Message) Format() string {
	return fmt.Sprintf("[%s][%s]: %s",
		msg.Timestamp.Format("2006-01-02 15:04:05"),
		msg.Sender.Name,
		msg.Content)
}

var clients = make(map[string]net.Conn)
var clientsMutex = &sync.Mutex{}

func StartServer(port string) {

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close() // we clean the port when done with program

	fmt.Println("Server listening on port" + port)

	// waiting for clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}
		fmt.Println("Waiting for connection")

		go HandleClient(conn, &clients, clientsMutex)
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
