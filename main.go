package main

import (
	"fmt"
	"netcat/handlers"
	"os"
	"strconv"
)

func main() {
	// * Get port from CLI args or default to 8989
	port := ":" + strconv.Itoa(handlers.GetPort())

	// * Start TCP server in a separate goroutine
	// ! If RunTCPServer blocks here, GUI will never start
	go func() {
		err := handlers.RunTCPServer(port)
		if err != nil {
			fmt.Println("Error starting server:", err)
			os.Exit(1) // ! Exit entire app if server fails
		}
	}()

	// * Launch admin GUI (blocks until closed)
	// ! If GUI fails, app exits immediately
	if err := handlers.RunGUI(); err != nil {
		fmt.Println("Error running GUI:", err)
		os.Exit(1)
	}
}
