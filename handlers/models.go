package handlers

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	Name       string
	Connection net.Conn
	Room       *Room
}

type Room struct {
	Name    string
	Members []*Client
	History []*Message
}

type Message struct {
	Timestamp time.Time
	Sender    *Client
	Content   string
}

func (msg Message) String() string {
	return fmt.Sprintf("[%s][%s]: %s",
		msg.Timestamp.Format("2006-01-02 15:04:05"),
		msg.Sender.Name,
		msg.Content)
}

func (room *Room) NewRoom() {
	room.Name = ""
	room.Members = nil
	room.History = nil
}
func (room *Room) SetName(name string) {
	room.Name = name
}

func (room *Room) AddMember(client Client) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()
	room.Members = append(room.Members, &client)
}

func (room *Room) RemoveMember(client Client) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()
	for i, member := range room.Members {
		if member.Name == client.Name {
			room.Members = append(room.Members[:i], room.Members[i+1:]...)
			break
		}
	}
}

func (room *Room) BroadcastMessage(message Message) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()
	room.History = append(room.History, &message)

	for _, member := range room.Members {
		fmt.Fprint(member.Connection, message.String())
	}
}