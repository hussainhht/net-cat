package handlers

import (
	"net"
	"time"
)

type Client struct {
	Name       string
	Connection net.Conn
	Room       Room
}

type Room struct {
	Name    string
	Members []Client
	History []Message
}

type Message struct {
	Timestamp time.Time
	Sender    Client
	Content   string
}