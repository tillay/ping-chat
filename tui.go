package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app     *tview.Application
	msgView map[string]*tview.TextView
	listNum int = 5
	activeChannel string = "channel 0"
)

func initTUI(onSend func(string)) {
	app = tview.NewApplication()
	pages := tview.NewPages()
	msgView = make(map[string]*tview.TextView)

	for i := 0; i < listNum; i++ {//initalize pages on per-channel basis
		name := fmt.Sprintf("channel %d", i)
		view, prim := channel(name, onSend)//channel returns TextView and Primitive
		msgView[name] = view
			pages.AddPage(name,prim,true,i == 0)
		}
	
	list := tview.NewList()//initalize list
	for i := 0; i < listNum; i++ {
		name := fmt.Sprintf("channel %d", i)
		list.AddItem(name,"",0, func(){
		activeChannel = name
		pages.SwitchToPage(name)
		})
	}

chat := tview.NewFlex().SetDirection(tview.FlexRow).
	AddItem(pages,0,1, true)
mainView := tview.NewFlex().
	AddItem(list,20,0,false).
	AddItem(chat,0,1,true)

	signIn := tview.NewForm().
		AddInputField("username", "", 20, nil, nil).
		AddInputField("password", "", 20, nil, nil).
		AddDropDown("username color",[]string{"red","blue","green"},0,nil).
		AddButton("enter", func() {
		username := signIn.GetFormItemByLabel("username").(*tview.InputField).GetText()
		password := signIn.GetFormItemByLabel("password").(*tview.InputField).GetText()
		color := signIn.GetDropDown("username color").(*tview.InputField).GetCurrentOption()
		fmt.println(username,password,color)
			pages.SwitchToPage("main")
		})
	signIn.SetBorder(true).SetTitle("enter details").SetTitleAlign(tview.AlignLeft)
	
	pages.AddPage("signIn",signIn,true,true)
	pages.AddPage("main",mainView,true,false)

app.SetRoot(pages,true).SetFocus(signIn) //puts the user in sign in to start
}

func channel(name string, onSend func(string)) (*tview.TextView, tview.Primitive){
	textView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true)
		textView.ScrollToEnd()
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetBorderPadding(0, 0, 1, 1)
	textView.SetBorder(true)
	textView.SetTitle(" PingChat v2 ")

		inputBox := tview.NewInputField()
	inputBox.SetBorder(true)
	inputBox.SetBackgroundColor(tcell.ColorDefault)
	inputBox.SetFieldBackgroundColor(tcell.ColorDefault)

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
		AddItem(textView, 0, 1, false).
		AddItem(inputBox, 3, 1, true)

	return textView, flex
}

func tuiPrint(line string) {
	view := msgView[activeChannel]
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(view, "%s\n", line)
	})
}

func runTUI() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func ScrollToMessage(){
	textView := msgView[activeChannel]
	app.QueueUpdateDraw(func(){
		textView.ScrollToEnd()
	})
}
