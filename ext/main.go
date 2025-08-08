package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

// Mock Types
type Client struct {
	Name string
}

type Room struct {
	Name    string
	Members []*Client
}

// Global vars
var (
	guiInstance        *gocui.Gui
	allRooms           []*Room
	selectedRoomIndex  int
	selectedUserIndex  int
)

func main() {
	// Mock data
	allRooms = []*Room{
		{Name: "General", Members: []*Client{{Name: "Bader"}, {Name: "Anwar"}}},
		{Name: "Gaming", Members: []*Client{{Name: "Batool"}, {Name: "Hussain"}}},
		{Name: "Coding", Members: []*Client{{Name: "Basel"}, {Name: "Taqi"}}},
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	guiInstance = g
	g.Cursor = false
	g.SetManagerFunc(layout)

	if err := setKeybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Rooms View
	v, err := g.SetView("rooms", 0, 0, 30, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Rooms"
	v.Highlight = true
	v.Clear()
	for i, room := range allRooms {
		prefix := "  "
		if i == selectedRoomIndex {
			prefix = "➤ "
		}
		fmt.Fprintf(v, "%s%s\n", prefix, room.Name)
	}

	// Users View
	v, err = g.SetView("users", 32, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Users in Room"
	v.Highlight = true
	v.Clear()
	if len(allRooms) > selectedRoomIndex {
		room := allRooms[selectedRoomIndex]
		for i, user := range room.Members {
			prefix := "  "
			if i == selectedUserIndex {
				prefix = "➤ "
			}
			fmt.Fprintf(v, "%s%s\n", prefix, user.Name)
		}
	}

	return nil
}

func setKeybindings(g *gocui.Gui) error {
	// Room navigation
	g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, prevRoom)
	g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, nextRoom)

	// User navigation
	g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, prevUser)
	g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, nextUser)

	// Kick selected user
	//g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, kickSelectedUser)

	// Quit
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)

	return nil
}

// Navigation functions
func nextRoom(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex < len(allRooms)-1 {
		selectedRoomIndex++
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func prevRoom(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex > 0 {
		selectedRoomIndex--
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func nextUser(g *gocui.Gui, v *gocui.View) error {
	room := allRooms[selectedRoomIndex]
	if selectedUserIndex < len(room.Members)-1 {
		selectedUserIndex++
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

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
