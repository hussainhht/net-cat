package handlers

import (
	"fmt"
	"log"
	"time"

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

	go refreshLayout()

	if err := setKeybindings(g); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	return nil
}

func refreshLayout() {
	// wati 1 second
	// for {
	// 	time.Sleep(1 * time.Second)
	// 	if guiInstance != nil {
	// 		guiInstance.Update(layout)
	// 		continue
	// 	}
	// }
	ticker := time.NewTicker(1 * time.Second) // Refresh every second
	defer ticker.Stop()

	for range ticker.C {
		if guiInstance != nil {
			guiInstance.Update(layout)
			continue
		}
	}

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
	v, err := g.SetView("rooms", 0, 0, maxX/3-1, maxY-15)
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


	// Room Stats view (below Rooms)
	if v, err = g.SetView("roomstats", 0, maxY-10, maxX/3-1, maxY-4); err != nil && err != gocui.ErrUnknownView {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Room Stats"
		v.Wrap = true
		v.Clear()
	}

	// Users view
	v, err = g.SetView("users", maxX/3, 0, maxX*2/3-1, maxY-15)
	if err != nil && err != gocui.ErrUnknownView {
		return err

	}
	v.Title = "Users in Room"
	v.Highlight = true
	v.Clear()

	if selectedRoomIndex < len(Rooms) && selectedRoomIndex >= 0 {
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

	// User Stats view (below User)
	if v, err = g.SetView("userstats", maxX/3, maxY-10, maxX*2/3-1, maxY-4); err != nil && err != gocui.ErrUnknownView {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "User Stats"
		v.Wrap = true
		v.Frame = true
		v.Clear()
	}

	// log view
	if v, err = g.SetView("log", maxX*2/3, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Log"
		v.Highlight = true
		v.Wrap = true
		v.Autoscroll = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorBlack

		// Print some log messages for demonstration
		fmt.Fprintln(v, "Log messages will appear here...")
	}

	// Footer
	if v, err := g.SetView("footer", 0, maxY-3, maxX*2/3-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		//v.Frame = false
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
	} else {
		selectedRoomIndex = len(Rooms) - 1 // Loop back to the first room
		selectedUserIndex = 0
	}
	g.Update(layout)
	
	return nil
}

func nextRoom(g *gocui.Gui, v *gocui.View) error {
	if selectedRoomIndex < len(Rooms)-1 {
		selectedRoomIndex++
		selectedUserIndex = 0
	} else {
		selectedRoomIndex = 0 // Loop back to the first room
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func prevUser(g *gocui.Gui, v *gocui.View) error {
	members := getRoomMembers(Rooms[selectedRoomIndex])
	if selectedUserIndex > 0 {
		selectedUserIndex--
	} else {
		selectedUserIndex = len(members)-1 // Stay at the first user
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
	} else {
		selectedUserIndex = 0 // Loop back to the first user
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
	logMsg(g, fmt.Sprintf("Kicked user: %s from room: %s", members[selectedUserIndex].Name, room.Name))	
	return nil
}

func logMsg(g *gocui.Gui,message string) {
	view, err := g.View("log")
	if err != nil {
		log.Println("Error getting log view:", err)
		return
	}
	fmt.Fprintln(view, message)
	view.Autoscroll = true 
	// fmt.Printf("Kicked user: %s\n", client.Name)
}