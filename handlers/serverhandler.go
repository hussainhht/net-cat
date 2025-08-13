package handlers

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const defaultPort string = ":8989"
const maxClients = 10

// Global client registry (guard with ClientsMutex)
var Clients []*Client            // store pointers to avoid copies
var ClientsMutex = &sync.Mutex{} // * Guards Clients and ClientsConnected

var Rooms []*Room
var RoomsMutex = &sync.Mutex{} // * Guards Rooms slice and room internals if needed

// * Starts TCP server and accepts clients
func RunTCPServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close() // * Ensure port is released on exit

	fmt.Printf("Listening on the port %s\n", port)

	// * Accept loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}
		// * Handle client in separate goroutine; registration will enforce max under lock
		go HandleClientConnection(conn)
	}
}
// * Parses CLI args and returns port (default :8989) or prints usage
func GetPort() string {
	args := os.Args

	if len(args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}
	port := defaultPort
	if len(args) == 2 {
		portnum, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("the port is not a number")
			os.Exit(1)
		}
		if portnum < 1024 || portnum > 65535 {
			fmt.Println("the port must be between 1024 and 65535")
			os.Exit(1)
		}
		port = ":" + args[1]
	}
	return port
}


// * Creates a new room and registers it globally
func CreateRoom(roomName string) *Room {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()

	newRoom := Room{
		Name:        roomName,
		Members:     make([]*Client, 0),
		History:     make([]*Message, 0),
		TimeCreated: time.Now(),
	}
	Rooms = append(Rooms, &newRoom)
	return &newRoom // ? Returns pointer to value stored in slice (OK as long as not reallocated frequently)
}
