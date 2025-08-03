package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to create listener, err:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Listening on %s\n", listener.Addr()) 

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection, err:", err)
			continue
		}
		go handleConnection(conn) 
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data, err:", err)
			}
			return
		}

		fmt.Printf("Request: %s", bytes)

		line := fmt.Sprintf("Echo %s", bytes)
		fmt.Printf("Response: %s", line) 

		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Failed to write data, err:", err)
			return
		}
	}
}
