package main

import (
	"net"
	"log"
	"netcat/handlers" // import your handler package that includes GuiHandler
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8989")
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Register your client manually (or receive init data from server)
	client := &handlers.Client{
		Name:       "Hussain",
		Connection: conn,
		// Room can be nil for now or initialized later based on server response
	}

	// Run the terminal GUI
	handlers.GuiHandler(client)
} 