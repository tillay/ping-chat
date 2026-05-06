package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app        *tview.Application
	msgView    *tview.TextView
	usersView  *tview.TextView
	statusView *tview.TextView
)

// dict of valid slash commands
// string name -> func to execute on command execution
// the input string to the function is any additional args after the command itself
var commands = map[string]func(args string){
	"/color": func(inputColor string) {
		go func() {
			*color = inputColor
			sendHandshake()
			tuiPrint("Set color to [" + inputColor + "]" + inputColor + "[white]")
		}()
	},
	"/name": func(inputName string) {
		go func() {
			*user = inputName
			sendHandshake()
			tuiPrint("Set name to " + inputName)
		}()
	},
	"/ping": func(_ string) {
		go func() {
			tuiPrint(sendInfoPing(*ip))
		}()
	},
	"/clear": func(_ string) {
		go func() {
			msgView.SetText("")
		}()
	},
}

func newView() *tview.TextView {
	v := tview.NewTextView().SetScrollable(true).SetDynamicColors(true).ScrollToEnd()
	v.SetBackgroundColor(tcell.ColorDefault)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetBorder(true)
	return v
}

func initTUI(onSend func(string)) {
	app = tview.NewApplication()
	app.EnableMouse(true)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		switch action {
		case tview.MouseLeftDown, tview.MouseLeftClick,
			tview.MouseRightDown, tview.MouseRightClick,
			tview.MouseMiddleDown, tview.MouseMiddleClick:
			return nil, action
		}
		return event, action
	})

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	msgView, usersView, statusView = newView(), newView(), newView()
	msgView.SetTitle(" PingChat v2 ")

	inputBox := tview.NewInputField()
	inputBox.SetBorder(true)
	inputBox.SetFieldBackgroundColor(tcell.ColorDefault)
	inputBox.SetLabelColor(tcell.ColorWhite)
	inputBox.SetLabel("> ")
	inputBox.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := inputBox.GetText()

		for command := range commands {
			if strings.HasPrefix(text, command) {
				commands[command](text[strings.Index(text, " ")+1:])
				inputBox.SetText("")
				return
			}
		}

		if text == "" || len(text) > 512 || isMsgOutgoing {
			return
		}
		inputBox.SetText("")
		isMsgOutgoing = true
		go onSend(text)
	})

	inputBox.SetAutocompleteStyles(
		tcell.ColorDefault,
		tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault),
		tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
	)

	inputBox.SetAutocompleteFunc(func(current string) []string {
		if !strings.HasPrefix(current, "/") {
			return nil
		}
		var matches []string
		if strings.HasPrefix(current, "/color") {
			for color := range tcell.ColorNames {
				if strings.HasPrefix("/color "+color, current) {
					matches = append(matches, "/color["+color+"] "+color)
				}
			}
		} else if strings.ReplaceAll(current, " ", "") == "/name" {
			matches = []string{"/name <new name>"}
		} else {
			for c := range commands {
				if strings.HasPrefix(c, current) {
					matches = append(matches, c)
				}
			}
		}
		return matches
	})

	inputBox.SetAutocompletedFunc(func(text string, index int, source int) bool {
		if source != tview.AutocompletedNavigate {
			if i := strings.Index(text, "] "); i != -1 {
				inputBox.SetText("/color " + text[i+2:])
			} else {
				inputBox.SetText(text)
			}
			return true
		}
		return false
	})

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(msgView, 0, 1, false).
		AddItem(inputBox, 3, 1, true)

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(usersView, 0, 1, false).
		AddItem(statusView, 3, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inner, 0, 1, true).
		AddItem(right, 0, 1, false)

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		w, _ := screen.Size()
		side := max(20, min(w/9, 32))
		flex.ResizeItem(right, side, 0)
		return false
	})
	setConnectedStatus(false)
	app.SetRoot(flex, true).SetFocus(inputBox)
}

func tuiPrint(line string) {
	// add a line to the scrolling text field
	if !*server {
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(msgView, "%s\n", line)
		})
	} else {
		// some parts of the low level code call this, so just print normally if it's running a server
		fmt.Println(line)
	}
}

func redrawUserView() {
	usersView.SetText("")
	for _, u := range onlineUsers {
		fmt.Fprintf(usersView, "[%s]%s[white]\n%s\n\n", u.Color, u.User, u.Loc)
	}
}

func setConnectedStatus(status bool) {
	if *server {
		return
	}
	statusText := "[green]✔[white] Connected"
	if status == false {
		statusText = "[red]✘[white] Not connected"
	}
	statusView.SetText(statusText)
}

func runTUI() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
