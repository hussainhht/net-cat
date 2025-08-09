package main

import (
	"fmt"
	"netcat/handlers"
	"os"
)

func main() {
	port := handlers.GetPort()

	// TCP server in a separate goroutine
	go func() {
		err := handlers.RunTCPServer(port)
		if err != nil {
			fmt.Println("Error starting server:", err)
			os.Exit(1)
		}
	}()

	// Admin GUI 
	if err := handlers.RunGUI(); err != nil {
		fmt.Println("Error running GUI:", err)
		os.Exit(1)
	}
}
