package main

import (
    "fmt"
    "net"
    "os"
)

func main() {
    
    listener, err := net.Listen("tcp", ":8989")
    if err != nil {
        fmt.Println("Error starting server:", err)
        os.Exit(1)
    }
    defer listener.Close()

    fmt.Println("Server listening on port 8989")

    // waiting for clients
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting client:", err)
            continue
        }

        fmt.Println("Client connected:", conn.RemoteAddr())
        conn.Close() // closing it for now 
    }
}