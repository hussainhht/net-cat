package handlers

import (
	"fmt"
	"strings"
	"time"
)

// * Sentinel error used to signal a clean client exit
var ErrClientExit = fmt.Errorf("client exit")

// * Parses slash commands and executes them
// Returns: (isCommand, error)
func HandleCommand(line string, client *Client) (bool, error) {
	line = strings.TrimSpace(line)
	if line == "" || line == "/" {
		return false, nil
	}

	pats := strings.SplitN(line, " ", 2) // ? args not used yet but kept for future
	cmd := strings.ToLower(pats[0])

	if !strings.HasPrefix(cmd, "/") {
		return false, nil
	}

	switch cmd {
	case "/help":
		return true, cmdHelp(client)
	case "/who", "/list":
		return true, cmdWho(client)
	case "/exit", "/quit":
		return true, cmdExit(client)
		case "/rename":
			return true, cmdRename(client, pats[1:])
	default:
		// * Return error for unknown commands
		if strings.HasPrefix(cmd, "/") {
			return true, client.Send(fmt.Sprintf("unknown command: %s", cmd))
		}
	}
	return true, nil
}

// * Sends a short help screen to the client
func cmdHelp(c *Client) error {
	help := `
Commands:
/help                 Show this help
/rename <newName>     Change your username
/who | /list          Show members in this room
/exit                 Leave the chat

`
	return c.Send(help + "\n") // * Send once instead of multiple writes
}

// * Lists members in the current room
func cmdWho(c *Client) error {
	r := c.Room
	if r == nil {
		return c.Send("You are not in a room.\n")
	}

	// ! Possible race: reading Clients without lock
	// * Better: read from r.Members under a room mutex
	var members []string
	RoomsMutex.Lock()
	for _, m := range r.Members {
		if m != nil {
			members = append(members, m.Name)
		}
	}
	RoomsMutex.Unlock()
	if len(members) == 0 {
		return c.Send("No members in this room.\n")
	}

	// ! Sending line-by-line can block
	// * Better: join and send once
	return c.Send(strings.Join(members, "\n") + "\n")
}

// ! not used (example for rename command with locking)
func cmdRename(c *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /rename <newName>")
	}
	newName := args[0]
	ClientsMutex.Lock() // * Protect global Clients list
	defer ClientsMutex.Unlock()
	if c.Name == newName {
		return c.Send("you already have that name\n")
	}
	for _, client := range Clients {
		if client.Name == newName {
			return c.Send(fmt.Sprintf("username '%s' is already taken\n", newName))
		}
	}
	old := c.Name
	c.Name = newName // * No need to clear before setting
	ms := fmt.Sprintf(" %s has been changed to %s and the old name was %s\n", old ,newName, old)
	c.Room.BroadcastMessage(Message{
		Timestamp: time.Now(),
		Sender:    &Client{Name: "SERVER"},
		Content:   ms,
	})
	return nil
}

// * Handles client exit: announce, remove, close connection
func cmdExit(c *Client) error {
	if c.Room != nil {
		leaveMsg := Message{
			Timestamp: time.Now(),
			Sender:    &Client{Name: "SERVER"}, // ? Consider using a singleton
			Content:   fmt.Sprintf("%s has left the chat.\n", c.Name),
		}
		c.Room.RemoveMember(c)
		c.Room.BroadcastMessage(leaveMsg)
	}

	ClientsMutex.Lock()
	for i, client := range Clients {
		if client == c || (client.Name == c.Name && client.Connection == c.Connection) {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}
	ClientsMutex.Unlock()

	_ = c.Connection.Close() // * Ignore error on close
	return ErrClientExit     // * Signal clean exit to the main loop
}
