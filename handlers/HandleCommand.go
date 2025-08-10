package handlers

import (
	"fmt"
	"strings"
	"time"
)

var ErrClientExit = fmt.Errorf("client exit")

func HandleCommand(line string, client *Client) (bool, error) {

	line = strings.TrimSpace(line)
	if line == "" || line == "/" {
		return false, nil
	}

	pats := strings.SplitN(line, " ", 2)
	cmd := strings.ToLower(pats[0])

	if !strings.HasPrefix(cmd, "/") {
		return false, nil
	}
	// args := pats[1:]

	switch cmd {
	case "/help":
		return true, cmdHelp(client)
	case "/who", "/list":
		return true, cmdWho(client)
	// case "/room":
	// 	return true, cmdRoom(client)
	// case "/rename":
	// 	return true, cmdRename(client, args)
	case "/exit", "/quit":
		return true, cmdExit(client)
	default:
		if strings.HasPrefix(cmd, "/") {
			return true, fmt.Errorf("unknown command: %s", cmd)
		}
	}
	return true, nil
}

func cmdHelp(c *Client) error {
	help := `
Commands:
/help                 Show this help
/who | /list          Show members in this room
/room                 Show your current room
/exit                 Leave the chat
`
	return c.Send(help + "\n")

}
func cmdWho(c *Client) error {
	r := c.Room
	if r == nil {
		return c.Send("You are not in a room.\n")
	}
	var members []string
	for _, client := range Clients {
		if client.Room == r {
			members = append(members, client.Name)
		}
	}
	if len(members) == 0 {
		return c.Send("No members in this room.\n")
	}
	for _, member := range members {
		c.Send(member + "\n")
	}
	return nil
}

func cmdRename(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /rename <newName>")
	}
	newName := args[0]

	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	if c.Name == newName {
		return c.Send("you already have that name")
	}
	for _, client := range Clients {
		if client.Name == newName {
			return c.Send(fmt.Sprintf("username '%s' is already taken", newName))

		}

	}
	old := c.Name

	c.Name = ""

	c.Name = newName
	return c.Send(fmt.Sprintf("Your name has been changed to %s\n and your old name was %s\n", newName, old))
}

func cmdExit(c *Client) error {
	if c.Room != nil {
		leaveMsg := Message{
			Timestamp: time.Now(),
			Sender:    &Client{Name: "SERVER"}, // virtual sender
			Content:   fmt.Sprintf("%s has left the chat.\n", c.Name),
		}
		c.Room.BroadcastMessage(leaveMsg)
		c.Room.RemoveMember(*c)
	}

	for i, client := range Clients {
		if client.Name == c.Name {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}

	c.Connection.Close()
	return ErrClientExit
}


