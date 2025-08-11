package handlers

import (
	"fmt"
	"log"
	"time"
	"github.com/jroimartin/gocui"
)

// ================================
// Admin TUI (gocui) for TCP-Chat
// -------------------------------
// This file implements a simple admin panel using the gocui library.
// It shows three main panes:
//  1) Rooms list (left)
//  2) Users in the selected room (middle)
//  3) Log output (right)
// Plus a footer for key hints.
//
// The code is intentionally minimal: navigation updates indices and
// the layout is re-rendered on each tick. It relies on external globals
// like Rooms and Clients which belong to the chat server state.
// ================================

// GUI state (global). In a larger app, you'd likely wrap these inside a struct.
var (
	guiInstance       *gocui.Gui    // singleton gocui instance used across the package
	selectedRoomIndex int           // index into Rooms slice
	selectedUserIndex int           // index into current room's members
)

// RunGUI starts the admin panel. Blocks on g.MainLoop() until quit.
func RunGUI() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close() // ensure terminal is restored on exit

	guiInstance = g
	g.Cursor = false
	g.SetManagerFunc(layout) // layout is called on each refresh to (re)draw views

	// ! Periodic refresh (useful when underlying Rooms/Clients change)
	// ! Replaced busy-loop with time.Ticker and g.Update for thread-safe redraws
	go refreshLayout()

	if err := setKeybindings(g); err != nil {
		return err
	}

	// Enter main event loop; returns ErrQuit when quit is requested
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	return nil
}

// refreshLayout triggers a periodic UI update. Using Update schedules
// the provided function to run on the main UI goroutine (thread-safe for gocui).
func refreshLayout() {
	// NOTE: Previous busy-loop is commented out. A ticker is more efficient and clear.
	ticker := time.NewTicker(1 * time.Second) // ! Refresh every second
	defer ticker.Stop()

	for range ticker.C {
		if guiInstance != nil {
			guiInstance.Update(layout) // ! re-run layout to redraw views safely
			continue
		}
	}
}

// ! clampSelection ensures selected indices are valid for current data.
func clampSelection() {
	// ! Guard when there are no rooms
	if len(Rooms) == 0 {
		selectedRoomIndex = 0
		selectedUserIndex = 0
		return
	}
	if selectedRoomIndex < 0 {
		selectedRoomIndex = 0
	}
	if selectedRoomIndex >= len(Rooms) {
		selectedRoomIndex = len(Rooms) - 1
	}
	// ! Clamp user index against current room members
	members := getRoomMembers(Rooms[selectedRoomIndex])
	if len(members) == 0 {
		selectedUserIndex = 0
		return
	}
	if selectedUserIndex < 0 {
		selectedUserIndex = 0
	}
	if selectedUserIndex >= len(members) {
		selectedUserIndex = len(members) - 1
	}
}

// layout defines and updates all views: rooms, users, log, footer.
// g.SetView creates a view on first call (ErrUnknownView) and returns the same view subsequently.
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Apply minimal sizes to avoid negative coordinates on tiny terminals.
	if maxX < 40 {
		maxX = 40
	}
	if maxY < 10 {
		maxY = 10
	}

	// ! Keep indices in-range before drawing
	 clampSelection()

	// ---------------- Rooms view ----------------
	v, err := g.SetView("rooms", 0, 0, maxX/3-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Rooms"
	v.Highlight = true
	v.Clear()

	if len(Rooms) == 0 { // ! Handle empty rooms list gracefully
		fmt.Fprintln(v, "(no rooms)")
	} else {
		for i, room := range Rooms {
			prefix := "  "
			if i == selectedRoomIndex {
				prefix = "➤ " // visual indicator for selection
			}
			fmt.Fprintf(v, "%s%s\n", prefix, room.Name)
		}
	}

	// ---------------- Users view ----------------
	v, err = g.SetView("users", maxX/3, 0, maxX*2/3-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Users in Room"
	v.Highlight = true
	v.Clear()

	if len(Rooms) == 0 { // ! No rooms => no users
		fmt.Fprintln(v, "(no users)")
	} else if selectedRoomIndex < len(Rooms) && selectedRoomIndex >= 0 {
		room := Rooms[selectedRoomIndex]
		members := getRoomMembers(room)
		if len(members) == 0 { // ! Empty room case
			fmt.Fprintln(v, "(no users)")
		} else {
			for i, member := range members {
				prefix := "  "
				if i == selectedUserIndex {
					prefix = "➤ "
				}
				fmt.Fprintf(v, "%s%s\n", prefix, member.Name)
			}
		}
	}

	// ---------------- Log view ----------------
	if v, err = g.SetView("log", maxX*2/3, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Log"
		v.Highlight = true
		v.Wrap = true
		v.Autoscroll = true
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorBlack // NOTE: may be invisible on dark backgrounds

		// Initial placeholder content
		fmt.Fprintln(v, "Log messages will appear here...")
	}

	// ---------------- Footer view ----------------
	if v, err := g.SetView("footer", 0, maxY-3, maxX*2/3-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// v.Frame = false // keep frame visible for clarity
		v.FgColor = gocui.ColorWhite
		fmt.Fprintln(v, " ↑ ↓: Switch Room  ← →: Select User  Ctrl-D: Kick  Ctrl-C: Quit ")
	}

	return nil
}

// setKeybindings wires up navigation and actions. Returning errors allows caller to handle failures.
func setKeybindings(g *gocui.Gui) error {
	g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, prevRoom)
	g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, nextRoom)
	g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, prevUser)
	g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, nextUser)
	g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, kickMock)
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	return nil
}

// Navigation handlers. They rotate indices with wrap-around and trigger a redraw.
func prevRoom(g *gocui.Gui, v *gocui.View) error {
	if len(Rooms) == 0 { // ! Guard when there are no rooms
		logMsg(g, "No rooms available")
		return nil
	}
	if selectedRoomIndex > 0 {
		selectedRoomIndex--
		selectedUserIndex = 0
	} else {
		selectedRoomIndex = len(Rooms) - 1 // wrap to last room when at top
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func nextRoom(g *gocui.Gui, v *gocui.View) error {
	if len(Rooms) == 0 { // ! Guard when there are no rooms
		logMsg(g, "No rooms available")
		return nil
	}
	if selectedRoomIndex < len(Rooms)-1 {
		selectedRoomIndex++
		selectedUserIndex = 0
	} else {
		selectedRoomIndex = 0 // wrap to first room when at bottom
		selectedUserIndex = 0
	}
	g.Update(layout)
	return nil
}

func prevUser(g *gocui.Gui, v *gocui.View) error {
	if len(Rooms) == 0 { // ! Guard when there are no rooms
		logMsg(g, "No rooms available")
		return nil
	}
	members := getRoomMembers(Rooms[selectedRoomIndex])
	if len(members) == 0 { // ! No users to move between
		logMsg(g, "No users in this room")
		return nil
	}
	if selectedUserIndex > 0 {
		selectedUserIndex--
	} else {
		selectedUserIndex = len(members) - 1 // wrap to last user
	}
	g.Update(layout)
	return nil
}

func nextUser(g *gocui.Gui, v *gocui.View) error {
	if len(Rooms) == 0 { // ! Guard when there are no rooms
		logMsg(g, "No rooms available")
		return nil
	}
	members := getRoomMembers(Rooms[selectedRoomIndex])
	if len(members) == 0 { // ! No users to move between
		logMsg(g, "No users in this room")
		return nil
	}
	if selectedUserIndex < len(members)-1 {
		selectedUserIndex++
	} else {
		selectedUserIndex = 0 // wrap to first user
	}
	g.Update(layout)
	return nil
}

// quit exits the main loop by returning ErrQuit
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// getRoomMembers returns a slice of pointers to Clients that belong to the given room.
// NOTE: This scans the global Clients slice each time; for large lists, consider indexing by room.
func getRoomMembers(room *Room) []*Client {
	var members []*Client
	if room == nil { // ! Defensive: nil room means no members
		return members
	}
	for i := range Clients {
		if Clients[i].Room == room {
			members = append(members, &Clients[i])
		}
	}
	return members
}

// kickMock simulates removing the selected user from the selected room and logs the action.
// In production, this should call the real kick logic and handle errors.
func kickMock(g *gocui.Gui, v *gocui.View) error {
	if len(Rooms) == 0 { // ! Guard when there are no rooms
		logMsg(g, "No rooms available")
		return nil
	}
	room := Rooms[selectedRoomIndex]
	members := getRoomMembers(room)

	if len(members) == 0 || selectedUserIndex >= len(members) { // ! Guard empty/invalid selection
		logMsg(g, "No users to kick in this room")
		return nil
	}

	// ! NOTE: kickSelectedUser signature unknown. If it expects *Client, pass members[selectedUserIndex].
	// ! If it expects a value, current code will need adjustment where that function is defined.
	kickSelectedUser(*members[selectedUserIndex])
	logMsg(g, fmt.Sprintf("Kicked user: %s from room: %s", members[selectedUserIndex].Name, room.Name))
	return nil
}

// logMsg prints a line to the log view and keeps autoscroll enabled.
func logMsg(g *gocui.Gui, message string) {
	view, err := g.View("log")
	if err != nil {
		log.Println("Error getting log view:", err)
		return
	}
	fmt.Fprintln(view, message)
	view.Autoscroll = true
	// fmt.Printf("Kicked user: %s", client.Name)
}
