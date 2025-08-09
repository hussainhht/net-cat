package handlers

import (
	"fmt"
	"log"
	// "math"

	"github.com/jroimartin/gocui"
)

// GUI state
var (
	guiInstance       *gocui.Gui
	selectedRoomIndex int
	selectedUserIndex int
)

// RunGUI starts the admin panel
func RunGUI() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	guiInstance = g
	g.Cursor = false
	g.SetManagerFunc(layout)

	if err := setKeybindings(g); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	return nil
}

// views: rooms, users, footer
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if maxX < 40 {
		maxX = 40
	}
	if maxY < 10 {
		maxY = 10
	}
	// Rooms view
	v, err := g.SetView("rooms", 0, 0, 30, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Rooms"
	v.Highlight = true
	v.Clear()

	for i, room := range Rooms {
		prefix := "  "
		if i == selectedRoomIndex {
			prefix = "➤ "
		}
		fmt.Fprintf(v, "%s%s\n", prefix, room.Name)
	}

	// Users view
	v, err = g.SetView("users", 32, 0, maxX-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Users in Room"
	v.Highlight = true
	v.Clear()

	if selectedRoomIndex < len(Rooms) {
		room := Rooms[selectedRoomIndex]
		members := getRoomMembers(room)

		for i, member := range members {
			prefix := "  "
			if i == selectedUserIndex {
				prefix = "➤ "
			}
			fmt.Fprintf(v, "%s%s\n", prefix, member.Name)
		}
	}

	// Footer
	if v, err := g.SetView("footer", 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.ColorBlue
		v.FgColor = gocui.ColorWhite
		fmt.Fprintln(v, " ↑ ↓: Switch Room  ← →: Select User  Ctrl-D: Kick  Ctrl-C: Quit ")
	}

	return nil
}

// Navigation & bindings
func setKeybindings(g *gocui.Gui) error {
	g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, prevRoom)
	g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, nextRoom)
	g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, prevUser)
	g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, nextUser)
	g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, kickMock)
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	return nil
}

// Navigation handlers
func prevRoom(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex > 0 {
		selectedRoomIndex--
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func nextRoom(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex < len(Rooms)-1 {
		selectedRoomIndex++
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func prevUser(g *gocui.Gui, v *gocui.View) error {
	if selectedUserIndex > 0 {
		selectedUserIndex--
	}
	g.Update(layout)
	return nil
}

func nextUser(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex < len(Rooms) {
		members := getRoomMembers(Rooms[selectedRoomIndex])
		if selectedUserIndex < len(members)-1 {
			selectedUserIndex++
		}
	}
	g.Update(layout)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}


// Yo this for getting users from a room
func getRoomMembers(room *Room) []*Client {
	var members []*Client
	for i := range Clients {
		if Clients[i].Room == room {
			members = append(members, &Clients[i])
		}
	}
	return members
}

// this is for keybinging 
func kickMock(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex >= len(Rooms) {
		return nil
	}
	room := Rooms[selectedRoomIndex]
	members := getRoomMembers(room)

	if selectedUserIndex >= len(members) {
		return nil
	}

	kickSelectedUser(*members[selectedUserIndex])

	return nil
}


