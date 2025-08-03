package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
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
		go handleConnection(conn)
	}
}

func requireName(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte("[" + time.Now().Format("2006-01-02 15:04:05") + "]"))
	conn.Write([]byte("Enter Username:"))
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read name")
		return "", err
	}
	name = strings.TrimSpace(name)
	return name, nil
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	username, usernameError := requireName(reader, conn)
	if usernameError != nil {
		fmt.Println(usernameError)
		return usernameError
	}
	fmt.Println("Connection Established from username: " + username)
	return	nil
}
