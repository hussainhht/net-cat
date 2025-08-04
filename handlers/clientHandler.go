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
	RegisterClient(clientsMutex, clients, username, conn)
	fmt.Println("Connection Established from username: " + username)
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

func RegisterClient(clientsMutex *sync.Mutex, clients *[]Client, username string, conn net.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range *clients {
		if client.Name == username {
			fmt.Fprint(conn, "Name already taken.\n")
			conn.Close()
			return
		}
	}

	newClient := Client{
		Name:       username,
		Connection: conn,
	}
	*clients = append(*clients, newClient)
}
