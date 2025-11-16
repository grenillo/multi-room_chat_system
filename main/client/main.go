package main

import (
	"multi-room_chat_system/client"
	//"fyne.io/fyne/v2"
	//"fyne.io/fyne/v2/app"
	//"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/widget"

	//"fyne.io/fyne/v2"
)

func main() {
	client.TestWindow()
	//client.StartClient()
	/*
	rooms := []string{"#general", "#staff"}

	a := app.New()
	w := a.NewWindow("Multi-Room Chat")
	w.Resize(fyne.NewSize(1200, 800))

	// --------------------------
	// LEFT: Room List
	// --------------------------
	listView := widget.NewList(
		func() int {
			return len(rooms)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(rooms[id])
		},
	)

	// --------------------------
	// CENTER: Chat History (scrollable)
	// --------------------------
	messagesBox := container.NewVBox()
	chatScroll := container.NewVScroll(messagesBox)
	chatScroll.SetMinSize(fyne.NewSize(600, 400))

	// --------------------------
	// BOTTOM: Input bar
	// --------------------------
	input := widget.NewEntry()
	input.SetPlaceHolder("Type a message...")

	sendBtn := widget.NewButton("Send", func() {
		text := input.Text
		if text != "" {
			// Add new message to box
			msg := widget.NewLabel("You: " + text)
			msg.Wrapping = fyne.TextWrapWord
			messagesBox.Add(msg)
			messagesBox.Refresh()

			// Auto-scroll to bottom
			chatScroll.ScrollToBottom()

			// Clear the input box
			input.SetText("")
		}
	})
	//enter key submits the message
	input.OnSubmitted = func(text string) {
		sendBtn.OnTapped()
	}

	// bottom bar: the TextEntry expands, the Button stays small
	bottomBar := container.NewBorder(nil, nil, nil, sendBtn, input)

	// --------------------------
	// RIGHT SIDE LAYOUT
	// --------------------------
	rightSide := container.NewBorder(nil, bottomBar, nil, nil, chatScroll)

	// --------------------------
	// HSplit: Room list + chat area
	// --------------------------
	split := container.NewHSplit(
		listView,
		rightSide,
	)
	split.Offset = 0.2

	w.SetContent(split)
	w.ShowAndRun()
	*/
}




