package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// ASCII welcome banner shown to every new connection
var ConnectionMessage string = "Welcome to TCP-Chat!\n" +
	"         _nnnn_\n" +
	"        dGGGGMMb\n" +
	"       @p~qp~~qMb\n" +
	"       M|@||@) M|\n" +
	"       @,----.JM|\n" +
	"      JS^\\__/  qKL\n" +
	"     dZP        qKRb\n" +
	"    dZP          qKKb\n" +
	"   fZP            SMMb\n" +
	"   HZM            MMMM\n" +
	"   FqM            MMMM\n" +
	" __| \".        |\\dS\"qML\n" +
	" |    `.       | `' \\Zq\n" +
	"_)      \\.___.,|     .'\n" +
	"\\____   )MMMMMP|   .'\n" +
	"     `-'       `--'\n\n"

// HandleClientConnection owns a single client connection lifecycle:
// 1) ask for username + room
// 2) register the client
// 3) stream messages until disconnect
func HandleClientConnection(conn net.Conn, clients *[]Client, clientsMutex *sync.Mutex) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// --- 1) Prompt for username (Anonymous if empty) ---
	username, usernameError := PromptForUsername(reader, conn)
	if usernameError != nil {
		fmt.Println(usernameError)
		return usernameError
	}

	// --- 2) Prompt for room (default if empty) ---
	requestedRoomName, roomError := PromptForRoom(reader, conn)
	if roomError != nil {
		fmt.Println(roomError)
		return roomError
	}

	// --- 3) Check if room exists, else create it ---
	var requestedRoom *Room
	for i := range Rooms {
		if Rooms[i].Name == requestedRoomName {
			requestedRoom = Rooms[i]
			break
		}
	}
	if requestedRoom == nil {
		requestedRoom = CreateRoom(requestedRoomName)
	}

	// --- 4) Register client (duplicate-name policy handled inside) ---
	client, registerError := RegisterClient(username, conn, requestedRoom)
	if registerError != nil {
		fmt.Println(registerError)
		return registerError
	}

	// --- 5) Send room history to the newcomer so they catch up ---
	DisplayRoomHistory(client)

	// --- 6) Main read/broadcast loop ---
	for {
		// Pre-print the prompt line with timestamp/name (UI polish)
		msg := Message{
			Timestamp: time.Now(),
			Sender:    client,
			Content:   "",
		}
		// Clear current terminal line and print the prefix
		fmt.Fprint(conn, "\r")
		fmt.Fprint(conn, msg)

		// Read one line from client
		msgContent, msgError := reader.ReadString('\n')

		if msgError != nil {

			// --- Client disconnected or network error: clean up ---
			client.Room.RemoveMember(*client)

			ClientsMutex.Lock()
			for i, c := range Clients {
				if c.Name == client.Name {
					Clients = append(Clients[:i], Clients[i+1:]...)
					break
				}
			}
			ClientsMutex.Unlock()

			// Inform the room that this client left
			leaveMsg := Message{
				Timestamp: time.Now(),
				Sender:    &Client{Name: "SERVER"}, // virtual/system sender
				Content:   fmt.Sprintf("%s has left the chat.\n", client.Name),
			}
			client.Room.BroadcastMessage(leaveMsg)
			return msgError
		}

		// Ignore whitespace-only messages (safer than checking just "\n")
		if strings.TrimSpace(msgContent) == "" {
			continue
		}

		// --- Commands (e.g., /exit, /rename, etc.) ---
		if consumed, err := HandleCommand(msgContent, client); consumed {
			if err != nil {
				fmt.Println("Error handling command:", err)
			}
			continue
		}

		// Wipe the "prompt" line we printed above
		fmt.Fprint(conn, "\033[1A")
		fmt.Fprint(conn, "\033[2K")

		// Broadcast the actual message
		msg = Message{
			Timestamp: time.Now(),
			Sender:    client,
			Content:   msgContent,
		}
		client.Room.BroadcastMessage(msg)

		// Update last activity
		clientsMutex.Lock()
		client.LastActive = time.Now()
		clientsMutex.Unlock()
	}
}

// PromptForUsername asks the user for a name.
// If empty or only whitespace, it assigns "Anonymous".
func PromptForUsername(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte(ConnectionMessage))
	conn.Write([]byte("[ENTER YOUR NAME]:"))

	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read name")
		return "", err
	}

	// Normalize whitespace first, then decide
	username = strings.TrimSpace(username)
	if username == "" {
		username = "Anonymous" // Default username if none specified
	}

	// if the name >15 program take only first 15 chars
	if len(username) > 15 {
		username = username[:15]
	}

	// Clean the prompt line in user's terminal
	fmt.Fprint(conn, "\033[1A") // Move cursor up one line
	fmt.Fprint(conn, "\033[2K")	// Clear the line
	return username, nil
}

// PromptForRoom asks the user for a room name.
// If empty, it assigns "default".
func PromptForRoom(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte("Enter Room:"))

	room, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read room")
		return "", err
	}

	// Normalize whitespace then fallback
	room = strings.TrimSpace(room)
	if room == "" {
		room = "default"
	}

	// Clean the prompt line in user's terminal
	fmt.Fprint(conn, "\033[1A")
	fmt.Fprint(conn, "\033[2K")
	return room, nil
}

// DisplayRoomHistory sends all previous messages in the room to the client.
func DisplayRoomHistory(client *Client) error {
	room := client.Room
	if room == nil {
		return fmt.Errorf("client is not in a room")
	}
	for _, msg := range room.History {
		fmt.Fprint(client.Connection, msg.String())
	}
	return nil
}

// RegisterClient adds a new client to global list + room,
// and applies duplicate-name policy:
// - If name is "Anonymous": allow multiples by appending a numeric suffix.
// - If name is custom and already taken: reject and close connection.
func RegisterClient(username string, conn net.Conn, room *Room) (*Client, error) {
	countname := 0

	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	orig := username
	for _, client := range Clients {
		if client.Name == username {
			countname++
			if orig == "Anonymous" {
				// For Anonymous users: Anonymous_1, Anonymous_2, ...
				username = fmt.Sprintf("%s_%d", orig, countname)
			} else {
				// Custom duplicate: reject
				fmt.Fprint(conn, "Username is already taken. Please choose a different name.\n")
				conn.Close()
				return nil, fmt.Errorf("name is already taken")
			}
		}
	}

	newClient := Client{
		Name:       username,
		Connection: conn,
		Room:       room,
	}

	// Add to global list
	Clients = append(Clients, newClient)

	// Inform the room that this client joined
	joinMsg := Message{
		Timestamp: time.Now(),
		Sender:    &Client{Name: "SERVER"}, // virtual/system sender
		Content:   fmt.Sprintf("%s has joined the chat.\n", newClient.Name),
	}
	newClient.Room.BroadcastMessage(joinMsg)

	// Add to room members
	room.AddMember(newClient)

	return &newClient, nil
}

// kickSelectedUser removes client from room and global list, closes the connection.
func kickSelectedUser(client Client) {

	client.Room.RemoveMember(client)

	for i, c := range Clients {
		if c.Name == client.Name {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}
	client.Connection.Close()
}
