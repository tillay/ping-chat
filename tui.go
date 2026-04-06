package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app        *tview.Application
	msgView    *tview.TextView
	usersView  *tview.TextView
	statusView *tview.TextView
)

func newView() *tview.TextView {
	v := tview.NewTextView().SetScrollable(true).SetDynamicColors(true).ScrollToEnd()
	v.SetBackgroundColor(tcell.ColorDefault)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetBorder(true)
	return v
}

func initTUI(onSend func(string)) {
	app = tview.NewApplication()
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
		if text == "" || len(text) > 512 {
			return
		}
		inputBox.SetText("")
		go onSend(text)
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

func userViewPrint(line string) {
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(usersView, "%s\n", line)
	})
}

func setConnectedStatus(status bool) {
	if *server {
		return
	}
	statusText := "[green]⬤[white] Connected"
	if status == false {
		statusText = "[red]⬤[white] Not connected"
	}
	statusView.SetText(statusText)
}

func runTUI() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
