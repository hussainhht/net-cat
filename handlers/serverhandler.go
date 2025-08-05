package handlers

import (
	"fmt"
	"net"
	"os"
	"sync"
)

const defaultPort string = ":8989"
const maxClients = 10

var Clients []Client
var ClientsMutex = &sync.Mutex{}

var Rooms []*Room
var RoomsMutex = &sync.Mutex{}

func RunTCPServer(port string) error {

	listener, err := net.Listen("tcp", port)

	if err != nil {
		return err
	}
	defer listener.Close() // we clean the port when done with program

	fmt.Printf("Listening on the port %s\n", port)

	// waiting for clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}

		// Yo Bader this is for checking if we've reached 10 Clients
		ClientsMutex.Lock()
		if len(Clients) >= maxClients {
			fmt.Fprint(conn, "Server is full. Maximum 10 clients allowed.\n")
			conn.Close()
			ClientsMutex.Unlock()
			continue
		}
		ClientsMutex.Unlock()

		go HandleClientConnection(conn, &Clients, ClientsMutex)
	}
}

func GetPort() string {
	args := os.Args
	if len(args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}

	port := defaultPort
	if len(args) == 2 {
		port = ":" + args[1]
	}
	return port
}

func CreateRoom(roomName string) *Room {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()

	newRoom := Room{
		Name:    roomName,
		Members: make([]*Client, 0),
		History: make([]*Message, 0),
	}
	Rooms = append(Rooms, &newRoom)
	return &newRoom // Return pointer to the room in the slice
}
