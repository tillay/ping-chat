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

func initTUI(onSend func(string)) {
	app = tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	// define the part where new messages scroll in
	msgView = tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		ScrollToEnd()

	msgView.SetBackgroundColor(tcell.ColorDefault)
	msgView.SetBorderPadding(0, 0, 1, 1)
	msgView.SetBorder(true)
	msgView.SetTitle(" PingChat v2 ")

	// define the box at the bottom where the user types
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

	usersView = tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		ScrollToEnd()
	usersView.SetBackgroundColor(tcell.ColorDefault)
	usersView.SetBorderPadding(0, 0, 1, 1)
	usersView.SetBorder(true)

	statusView = tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		ScrollToEnd()
	statusView.SetBackgroundColor(tcell.ColorDefault)
	statusView.SetBorderPadding(0, 0, 1, 1)
	statusView.SetBorder(true)

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(msgView, 0, 1, false).
		AddItem(inputBox, 3, 1, true)

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(usersView, 0, 1, false).
		AddItem(statusView, 3, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inner, 0, 1, true).
		AddItem(right, 0, 1, false)

	// this nicely adjusts the proportions of the messages vs users box
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		side := width / 9
		if side < 20 {
			side = 20
		}
		if side > 32 {
			side = 32
		}
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

func sideViewPrint(line string) {
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(usersView, "%s\n", line)
	})
}

func setConnectedStatus(status bool) {
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
