package handlers

import (
	"fmt"
	"strings"
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
	// case "rename":
	// 	return true, cmdRename(client, args)
	// case "/exit", "/quit":
	// 	return true, cmdExit(client)
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
/rename <newName>     Change your username
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

func (c *Client) Send(s string) error {
	_, err := fmt.Fprint(c.Connection, s)

	return err

}
