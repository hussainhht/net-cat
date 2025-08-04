package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
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
	registerError := RegisterClient(username, conn, requestedRoom)
	if registerError != nil {
		fmt.Println(registerError)
		return registerError
	}
	fmt.Println("Connection Established from username: " + username)
	fmt.Printf("Room '%s' now has %d members: %v\n", requestedRoom.Name, len(requestedRoom.Members), requestedRoom.Members)
	
	fmt.Println(Rooms)
	for _, client := range Rooms[0].Members {
		fmt.Println(client.Name)
	}
	return nil
}

func PromptForUsername(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte(ConnectionMessage))
	conn.Write([]byte("Enter Username:"))
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read name")
		return "", err
	}
	username = strings.TrimSpace(username)
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
	return room, nil
}

func RegisterClient(username string, conn net.Conn, room *Room) error {
	// First, check if username is taken and add client
	ClientsMutex.Lock()
	for _, client := range Clients {
		if client.Name == username {
			ClientsMutex.Unlock()
			conn.Close()
			return fmt.Errorf("name is already taken")
		}
	}

	newClient := Client{
		Name:       username,
		Connection: conn,
		Room:       room,
	}
	Clients = append(Clients, newClient)
	room.AddMember(newClient)
	return nil
}
