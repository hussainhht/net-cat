package main

import (
	"fmt"
	"log"
	"strings"
	"github.com/jroimartin/gocui"
)

var (
	username = "Batool"
	roomName = "Real Madrid On Top"
	users    = []string{"Bader", "Hussain","Batool"}
	messages = []string{
		"[2025-08-04 10:00:00][Bader]: Hello everyone!",
		"[2025-08-04 10:02:30][Batool]: Hey folks, good to see you.",
	}
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size() // terminal size

	// Room name top
	if v, err := g.SetView("room", 0, 0, maxX-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = "Current Room"
		v.Clear()
		fmt.Fprintf(v, " Room: %s ", roomName)
	}

	// Users list right
	if v, err := g.SetView("users", maxX-25, 3, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Users"
		v.Clear()
		for _, u := range users {
			fmt.Fprintf(v, "• %s\n", u)
		}
	}

	// Chat messages area left side
	if v, err := g.SetView("chat", 0, 3, maxX-26, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chat"
		v.Wrap = true
		v.Autoscroll = true
		v.Clear()
		for _, m := range messages {
			fmt.Fprintln(v, m)
		}
	}

	// Input box at bottom
	if v, err := g.SetView("input", 0, maxY-5, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf("Message (%s)", username)
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
	}

	// Footer with commands/help
	if v, err := g.SetView("footer", 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Clear()
		fmt.Fprintln(v, "Commands: /rename <newname> | /room <roomname> | Ctrl+C to quit")
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		input := strings.TrimSpace(v.Buffer())
		if input != "" {
			messages = append(messages, fmt.Sprintf("[2025-08-04 10:05:00][%s]: %s", username, input))
			v.Clear()
			v.SetCursor(0, 0)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
