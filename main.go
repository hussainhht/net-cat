package main

import (
	"netcat/handlers"
)

func main() {
	port := handlers.GetPort()
	handlers.RunTCPServer(port)
}
