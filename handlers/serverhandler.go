package handlers

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const defaultPort string = ":8989"
const maxClients = 10

var ClientsConnected = 0            // ! Keep in sync with Clients; prefer deriving from len(Clients)
var Clients []Client                // ! Stores by value; consider []*Client to avoid copies
var ClientsMutex = &sync.Mutex{}    // * Guards Clients and ClientsConnected

var Rooms []*Room
var RoomsMutex = &sync.Mutex{}      // * Guards Rooms slice and room internals if needed

// * Starts TCP server and accepts clients
func RunTCPServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close() // * Ensure port is released on exit

	fmt.Printf("Listening on the port %s\n", port)

	go updateClientCount() // ? Background cleanup; see notes inside for pitfalls

	// * Accept loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}

		// * Increment connected count
		ClientsMutex.Lock()
		ClientsConnected++
		ClientsMutex.Unlock()

		// * Enforce max connections
		ClientsMutex.Lock()
		if ClientsConnected > maxClients {
			fmt.Fprint(conn, "Server is full. Maximum 10 clients allowed.\n")
			conn.Close()
			ClientsConnected--       // ! Keep in sync on early close
			ClientsMutex.Unlock()
			continue
		}
		ClientsMutex.Unlock()

		// * Handle client in separate goroutine
		go HandleClientConnection(conn, &Clients, ClientsMutex) // ? Passes slice by pointer; ensure handler uses the mutex for all access
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
		port = ":" + args[1]
	}
	return port
}

// * Creates a new room and registers it globally
func CreateRoom(roomName string) *Room {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()

	newRoom := Room{
		Name:    roomName,
		Members: make([]*Client, 0),
		History: make([]*Message, 0),
	}
	Rooms = append(Rooms, &newRoom)
	newRoom.TimeCreated = time.Now() // * Track creation time
	return &newRoom                  // ? Returns pointer to value stored in slice (OK as long as not reallocated frequently)
}

// * Periodically prunes disconnected clients
func updateClientCount() {
	ticker := time.NewTicker(1 * time.Second)
	// ! This approach attempts to Read 0 bytes which typically blocks;
	// ! consider using deadlines, a ping/pong, or checking write errors in Send/Broadcast.
	// * Alternative: maintain Clients as []*Client and remove on write/read error paths immediately.

	for range ticker.C {
		for _, client := range Clients {
			// ! This read will block unless connection is closed with an error.
			// ! Better: use SetReadDeadline with short timeout or a heartbeat.
			_, err := client.Connection.Read([]byte{0})
			if err != nil {
				ClientsMutex.Lock()
				for i, c := range Clients {
					if c.Name == client.Name { // ? Name-based match; pointer equality is more robust
						Clients = append(Clients[:i], Clients[i+1:]...)
						ClientsConnected-- // ! Keep counters consistent
						break
					}
				}
				ClientsMutex.Unlock()
			}
		}
	}
}
