package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

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

func HandleClientConnection(conn net.Conn, clients *[]Client, clientsMutex *sync.Mutex) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	username, usernameError := PromptForUsername(reader, conn)
	if usernameError != nil {
		fmt.Println(usernameError)
		return usernameError
	}

	requestedRoomName, roomError := PromptForRoom(reader, conn)
	if roomError != nil {
		fmt.Println(roomError)
		return roomError
	}
	var requestedRoom *Room
	// check if room already exists

	for i := range Rooms {
		if Rooms[i].Name == requestedRoomName {
			// room exists - connect client to room
			requestedRoom = Rooms[i] // Get pointer to the actual room in slice
			break
		}
	}
	// room doesnt already exists
	if requestedRoom == nil {
		requestedRoom = CreateRoom(requestedRoomName)
		fmt.Println("Creating new room")
	}
	client, registerError := RegisterClient(username, conn, requestedRoom)
	if registerError != nil {
		fmt.Println(registerError)
		return registerError
	}
	fmt.Println("Connection Established from username: " + username)
	fmt.Printf("Room '%s' now has %d members: %v\n", requestedRoom.Name, len(requestedRoom.Members), requestedRoom.Members)

	fmt.Println(Rooms)

	// Send join notification to all clients in the room
	joinMsg := Message{
		Timestamp: time.Now(),
		Sender:    &Client{Name: "SERVER"},
		Content:   fmt.Sprintf("%s has joined our chat...\n", client.Name),
	}
	client.Room.BroadcastMessage(joinMsg)

	// Launch GUI after successful registration
	fmt.Fprint(conn, "Launching GUI interface...\n")
	GuiHandler(client)
	return nil
}

func PromptForUsername(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte(ConnectionMessage))
	conn.Write([]byte("[ENTER YOUR NAME]:"))
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read name")
		return "", err
	}
	username = strings.TrimSpace(username)
	fmt.Fprint(conn, "\033[1A")
	fmt.Fprint(conn, "\033[2K")
	return username, nil
}

func PromptForRoom(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte("Enter Room:"))
	room, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read room")
		return "", err
	}
	room = strings.TrimSpace(room)
	fmt.Fprint(conn, "\033[1A")
	fmt.Fprint(conn, "\033[2K")
	return room, nil
}

func RegisterClient(username string, conn net.Conn, room *Room) (*Client, error) {
	// First, check if username is taken and add client
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()
	for _, client := range Clients {
		if client.Name == username {
			// ClientsMutex.Unlock() defer is used to ensure the mutex is released
			fmt.Fprint(conn, "Username is already taken. Please choose a different name.\n")
			conn.Close()
			return nil, fmt.Errorf("name is already taken")
		}
	}

	newClient := Client{
		Name:       username,
		Connection: conn,
		Room:       room,
	}
	Clients = append(Clients, newClient)
	room.AddMember(newClient)
	return &newClient, nil
}
