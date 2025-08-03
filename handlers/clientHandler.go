package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	// "time"
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

func HandleClient(conn net.Conn, clients *map[string]net.Conn, clientsMutex *sync.Mutex) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	username, usernameError := requireName(reader, conn)
	if usernameError != nil {
		fmt.Println(usernameError)
		return usernameError
	}
	addClient(clientsMutex, clients, username, conn)
	fmt.Println("Connection Established from username: " + username)
	return nil
}

func requireName(reader *bufio.Reader, conn net.Conn) (string, error) {
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

func addClient(clientsMutex *sync.Mutex, clients *map[string]net.Conn, username string, conn net.Conn) {
	clientsMutex.Lock()
	if _, exists := (*clients)[username]; exists {
		clientsMutex.Unlock()
		fmt.Fprint(conn, "Name already taken.\n")
		conn.Close()
		return
	}
	(*clients)[username] = conn
	clientsMutex.Unlock()
	fmt.Println("unlocked")
}
