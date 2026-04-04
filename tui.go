package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app     *tview.Application
	msgView *tview.TextView
)

func initTUI(onSend func(string)) {
	app = tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	msgView = tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		ScrollToEnd()

	msgView.SetBackgroundColor(tcell.ColorDefault)
	msgView.SetBorderPadding(0, 0, 1, 1)
	msgView.SetBorder(true)
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

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(msgView, 0, 1, false).
		AddItem(inputBox, 3, 1, true)

	app.SetRoot(flex, true).SetFocus(inputBox)
}

func tuiPrint(line string) {
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(msgView, "%s\n", line)
	})
}

func runTUI() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
