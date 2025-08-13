package handlers

import (
	"fmt"
	"net"
	"time"
)

// * Represents a connected user
type Client struct {
	Name       string    // Client's display name
	Connection net.Conn  // TCP connection to this client
	Room       *Room     // Current room the client is in
	LastActive time.Time // Last time client sent/received activity
}

// * Represents a chat room
type Room struct {
	Name        string     // Room name
	Members     []*Client  // List of members (guarded by RoomsMutex)
	History     []*Message // Stored messages for history sync (guarded by RoomsMutex)
	TimeCreated time.Time  // When the room was created
}

// * Represents a single chat message
type Message struct {
	Timestamp time.Time // Time the message was sent
	Sender    *Client   // Who sent the message
	Content   string    // The message text
}

// * Formats a message into the required string format
func (msg Message) String() string {
	return fmt.Sprintf("[%s][%s]:%s",
		msg.Timestamp.Format("2006-01-02 15:04:05"),
		msg.Sender.Name,
		msg.Content)
}

// * Resets room to an empty state
func (room *Room) NewRoom() {
	room.Name = ""
	room.Members = nil
	room.History = nil
}

// * Sets the room's name
func (room *Room) SetName(name string) {
	room.Name = name
}

// * Adds a member to the room (thread-safe)
func (room *Room) AddMember(client *Client) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()
	// prevent duplicates
	for _, m := range room.Members {
		if m == client {
			return
		}
	}
	room.Members = append(room.Members, client)
}

// * Removes a member from the room by matching name (thread-safe)
func (room *Room) RemoveMember(client *Client) {
	RoomsMutex.Lock()
	defer RoomsMutex.Unlock()
	for i, member := range room.Members {
		if member == client || member.Name == client.Name {
			room.Members = append(room.Members[:i], room.Members[i+1:]...)
			break
		}
	}
}

// * Broadcasts a message to all members of the room (thread-safe)
// ? History is stored for new clients when they join
func (room *Room) BroadcastMessage(message Message) {
	if room == nil {
		return // ! Avoid nil pointer if room is missing
	}
	RoomsMutex.Lock()
	// Store message in history (keep only last N if needed)
	room.History = append(room.History, &message)
	// members := append([]*Client(nil), room.Members...) // copy for safe iteration
	RoomsMutex.Unlock()

	// Send to all members
	for _, member := range room.Members {
		fmt.Fprint(member.Connection, "\r") // Clear current line in terminal
		fmt.Fprint(member.Connection, message.String())
		fmt.Fprint(member.Connection, Message{
			Timestamp: time.Now(),
			Sender:    member,
			Content:   "",
		})
	}
}

// * Sends a raw string directly to the client connection
func (c *Client) Send(s string) error {
	_, err := fmt.Fprint(c.Connection, s)
	return err
}
