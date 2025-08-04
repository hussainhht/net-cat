package main

import (
	"fmt"
	"netcat/handlers"
	"os"
)

func main() {
	port := handlers.GetPort()

	fmt.Printf("Starting TCP Chat Server on port %s\n", port)

	err := handlers.RunTCPServer(port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
