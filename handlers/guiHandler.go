package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

// GUI state variables filled with Client Data
var (
	currentUsername string
	currentRoomName string
	currentUsers    []string
	currentMessages []string
	guiInstance     *gocui.Gui
	clientInstance  *Client
)

func GuiHandler(client *Client) {
	clientInstance = client

	// Getting GUI variables from client data
	updateGUIVariables()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	guiInstance = g

	g.Cursor = true
	g.SetManagerFunc(layout)

	if err := keybindings(g, client); err != nil {
		log.Panicln(err)
	}

	// Start a goroutine to refresh the GUI periodically
	go refreshGUI()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func updateGUIVariables() {
	if clientInstance == nil {
		return
	}

	currentUsername = clientInstance.Name
	if clientInstance.Room != nil {
		currentRoomName = clientInstance.Room.Name

		// Get users from room members
		currentUsers = []string{}
		for _, member := range clientInstance.Room.Members {
			currentUsers = append(currentUsers, member.Name)
		}

		// Get messages from room history
		currentMessages = []string{}
		for _, msg := range clientInstance.Room.History {
			currentMessages = append(currentMessages, msg.String())
		}
	}
}

func refreshGUI() {
	ticker := time.NewTicker(1 * time.Second) // Refresh every second
	defer ticker.Stop()

	for range ticker.C {
		if guiInstance != nil {
			updateGUIVariables()
			guiInstance.Update(func(g *gocui.Gui) error {
				return nil
			})
		}
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size() // terminal size

	// Room name top
	v, err := g.SetView("room", 0, 0, maxX-1, 2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = "Current Room"
	}
	v.Clear()
	fmt.Fprintf(v, " Room: %s ", currentRoomName)

	// Users list right
	v, err = g.SetView("users", maxX-25, 3, maxX-1, maxY-5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Users"
	}
	v.Clear()
	for _, u := range currentUsers {
		fmt.Fprintf(v, "• %s\n", u)
	}

	// Chat messages area left side
	v, err = g.SetView("chat", 0, 3, maxX-26, maxY-5)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chat"
		v.Wrap = true
		v.Autoscroll = true
	}
	v.Clear()
	for _, m := range currentMessages {
		fmt.Fprintln(v, m)
	}

	// Input box at bottom
	v, err = g.SetView("input", 0, maxY-5, maxX-1, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf("Message (%s)", currentUsername)
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
	}

	// Footer with commands/help
	v, err = g.SetView("footer", 0, maxY-3, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	v.Clear()
	fmt.Fprintln(v, "Commands: /rename <newname> | /room <roomname> | Ctrl+C to quit")

	return nil
}

func keybindings(g *gocui.Gui, client *Client) error {

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		input := strings.TrimSpace(v.Buffer())
		// Don't broadcast empty messages
		if input != "" {
			// Create a new message
			msg := Message{
				Timestamp: time.Now(),
				Sender:    client,
				Content:   input,
			}

			// Try to broadcast the message to all clients in the room
			if client.Room != nil {
				err := broadcastMessage(msg)
				if err != nil {
					// If broadcasting fails, still add to local display
					log.Printf("Failed to broadcast message: %v", err)
				}
			}

			// Add to local messages for GUI display
			currentMessages = append(currentMessages, msg.String())
		}

		v.Clear()
		v.SetCursor(0, 0)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func broadcastMessage(msg Message) error {
	if clientInstance == nil || clientInstance.Room == nil {
		return fmt.Errorf("client or room is nil")
	}

	// Add message to room history
	clientInstance.Room.History = append(clientInstance.Room.History, &msg)

	// Broadcast to all clients in the room
	for _, member := range clientInstance.Room.Members {
		if member.Connection != nil {
			_, err := fmt.Fprintf(member.Connection, "%s\n", msg.String())
			if err != nil {
				log.Printf("Failed to send message to %s: %v", member.Name, err)
			}
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	// Clean up when quitting
	if clientInstance != nil && clientInstance.Room != nil {
		// Remove client from room
		clientInstance.Room.RemoveMember(*clientInstance)

		// Remove from global clients list
		ClientsMutex.Lock()
		for i, c := range Clients {
			if c.Name == clientInstance.Name {
				Clients = append(Clients[:i], Clients[i+1:]...)
				break
			}
		}
		ClientsMutex.Unlock()

		// Send leave message
		leaveMsg := Message{
			Timestamp: time.Now(),
			Sender:    &Client{Name: "SERVER"},
			Content:   fmt.Sprintf("%s has left the chat.\n", clientInstance.Name),
		}
		broadcastMessage(leaveMsg)
	}

	return gocui.ErrQuit
}
