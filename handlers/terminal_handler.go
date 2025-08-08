package handlers

import (
	"bufio"
	"fmt"
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
		msg.Content = input
	}

}
