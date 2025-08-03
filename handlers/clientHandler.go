package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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

func HandleClient(conn net.Conn) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	username, usernameError := requireName(reader, conn)
	if usernameError != nil {
		fmt.Println(usernameError)
		return usernameError
	}
	fmt.Println("Connection Established from username: " + username)
	return nil
}

func requireName(reader *bufio.Reader, conn net.Conn) (string, error) {
	conn.Write([]byte(ConnectionMessage))
	conn.Write([]byte("Enter Username:"))
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("couldnt read name")
		return "", err
	}
	name = strings.TrimSpace(name)
	return name, nil
}
