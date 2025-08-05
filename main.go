package main

import (
	"fmt"
	"netcat/handlers"
	"os"
)

func main() {
	port := handlers.GetPort()

	err := handlers.RunTCPServer(port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
