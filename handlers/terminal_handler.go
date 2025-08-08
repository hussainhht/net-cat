package handlers

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func TerminalHandler(client *Client, reader *bufio.Reader) {
	conn := client.Connection

	fmt.Fprint(conn, "Welcome to the terminal interface!\n")

	for {
		msg := Message{
			Timestamp: time.Now(),
			Sender:    client,
			Content:   "",
		}
		fmt.Fprint(conn, msg)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprint(conn, "Error reading input, please try again.\n")
			continue
		}
		msg.Content = strings.TrimSpace(input)
		switch msg.Content {
		case "/exit":
			fmt.Fprint(conn, "Exiting terminal interface...\n")
			cleanupClient(client)

			return
		case "/help":
			fmt.Fprint(conn, "Available commands:\n")
			fmt.Fprint(conn, " - /exit: Exit the terminal interface\n")
			fmt.Fprint(conn, " - /help: Show this help message\n")
			fmt.Fprint(conn, " - /rename: Change your username\n")
			fmt.Fprint(conn, " - /room: Change your room\n")
		case "/rename":
			fmt.Fprint(conn, "Enter new username: ")
			newName, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprint(conn, "Error reading new username, please try again.\n")
				continue
			}
			newName = strings.TrimSpace(newName)
			client.Name = newName
			fmt.Fprintf(conn, "Username changed to: %s\n", newName)

		case "/who":
			if client.Room == nil {
				fmt.Fprint(conn, "You are not in any room.\n")
			} else {
				fmt.Fprintf(conn, "Members in room %s:\n", client.Room.Name)
				for _, member := range client.Room.Members {
					fmt.Fprintf(conn, "- %s\n", member.Name)
				}
			}

		case "/room":
			fmt.Fprint(conn, "Enter new room name: ")
			newRoom, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprint(conn, "Error reading new room name, please try again.\n")
				continue
			}
			newRoom = strings.TrimSpace(newRoom)
			if newRoom == "" {
				fmt.Fprint(conn, "Room name cannot be empty.\n")
				continue
			}
			switchRoom(client, newRoom)

		case "/rooms":
			if len(Rooms) == 0 {
				fmt.Fprint(conn, "No rooms available.\n")
			} else {
				fmt.Fprint(conn, "Available rooms:\n")
				for _, room := range Rooms {
					fmt.Fprintf(conn, "- %s\n", room.Name)
				}
			}
		default:
			msg := Message{
				Timestamp: time.Now(),
				Sender:    client,
				Content:   input,
			}
			if client.Room != nil {
				client.Room.BroadcastMessage(msg)
			} else {
				fmt.Fprint(conn, "You are not in a room. Use /room <name>\n")
			}

		}
	}

}

func switchRoom(client *Client, roomName string) {
	if client.Room != nil {
		client.Room.RemoveMember(*client)
	}

	var target *Room
	for i := range Rooms {
		if Rooms[i].Name == roomName {
			target = Rooms[i]
			break
		}
	}
	if target == nil {
		target = CreateRoom(roomName)
	}

	client.Room = target
	target.AddMember(*client)

	join := Message{
		Timestamp: time.Now(),
		Sender:    &Client{Name: "SERVER"},
		Content:   fmt.Sprintf("%s joined room %s\n", client.Name, roomName),
	}
	target.BroadcastMessage(join)
}

func cleanupClient(client *Client) {
	if client.Room != nil {
		client.Room.RemoveMember(*client)
	}

	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	for i, c := range Clients {
		if c.Name == client.Name {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}

	leave := Message{
		Timestamp: time.Now(),
		Sender:    &Client{Name: "SERVER"},
		Content:   fmt.Sprintf("%s has left the chat.\n", client.Name),
	}
	if client.Room != nil {
		client.Room.BroadcastMessage(leave)
	}
}
