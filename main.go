package main

import (
	"fmt"
	"net"
	"os"

	"netcat/handlers"
)

func main() {
	Args := os.Args
	if len(Args) > 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}
	port := ":8989"
	if len(Args) == 2 {
		port = ":"+Args[1]
	}
	
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
		go handlers.HandleClient(conn)
	}
}



