package client

import (
	//"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

var loginWin fyne.Window

type GUI struct {
	quitting 	bool
    roomBoxes   map[string]*fyne.Container
    chatScrolls map[string]*container.Scroll
	lobbyBox 	*fyne.Container
	lobbyScroll *container.Scroll
    inputChan   chan string
    window      fyne.Window
    currentRoom string
	listView    *widget.List
	rooms		[]string
	bottomBar	*fyne.Container
	selectedID  widget.ListItemID
}


func (g *GUI) Display(room string, text string) {
	if g.quitting {
		return
	}
    var box *fyne.Container
    var scroll *container.Scroll

    // Decide which container to append to
    if room == "" {
        box = g.lobbyBox
        scroll = g.lobbyScroll
    } else {
        b, ok := g.roomBoxes[room]
        if !ok {
            b = container.NewVBox()
            s := container.NewVScroll(b)
            s.SetMinSize(fyne.NewSize(600, 400))
            g.roomBoxes[room] = b
            g.chatScrolls[room] = s
        }
        box = g.roomBoxes[room]
        scroll = g.chatScrolls[room]
    }

    // Just append the message
    label := widget.NewLabel(text)
    label.Wrapping = fyne.TextWrapWord
    box.Add(label)
    box.Refresh()
    scroll.ScrollToBottom()
}

func (g *GUI) DisplayImage(room string, url string) {
    box, ok := g.roomBoxes[room]
    if !ok {
        box = container.NewVBox()
        g.roomBoxes[room] = box
        g.chatScrolls[room] = container.NewVScroll(box)
    }

    // Ensure the URL is valid
    uri := storage.NewURI(url)
    if uri == nil {
        log.Println("Invalid URI for image:", url)
        box.Add(widget.NewLabel("Could not load image: " + url))
        box.Refresh()
        return
    }

    img := canvas.NewImageFromURI(uri)
    img.FillMode = canvas.ImageFillContain
    img.SetMinSize(fyne.NewSize(200, 200))

    box.Add(img)
    box.Refresh()
    if scroll, ok := g.chatScrolls[room]; ok {
        scroll.Refresh()
        scroll.ScrollToBottom()
    }
}



func (g *GUI) ClearRoom(room string) {
    if b, ok := g.roomBoxes[room]; ok {
        b.Objects = nil
        b.Refresh()
    }
    if s, ok := g.chatScrolls[room]; ok {
        s.Refresh()
    }
}

func (g *GUI) ClearLobby() {
	g.lobbyBox.RemoveAll()
    g.lobbyBox.Refresh()
    g.lobbyScroll.Refresh()
}

func (g *GUI) SelectRoom(room string) {
	g.currentRoom = room
    if g.listView != nil {
        for i, r := range g.rooms {
            if r == room {
                g.listView.Select(i)
                g.selectedID = i
                break
            }
        }
    }
}

func (g *GUI) DeselectRoom() {
	g.currentRoom = ""
    if g.listView != nil {
        g.listView.Unselect(g.selectedID)
        g.selectedID = -1
    }
}


func (g *GUI) SetRooms(rooms []string) {
    g.rooms = rooms
	//initialize per-room containers
	for _, room := range rooms {
		box := container.NewVBox()
        scroll := container.NewVScroll(box)
        scroll.SetMinSize(fyne.NewSize(600, 400))
        g.roomBoxes[room] = box
        g.chatScrolls[room] = scroll
	}
    g.listView.Refresh()
}

func (g *GUI) AddRoom(room string) {
    g.rooms = append(g.rooms, room)
	//initialize room containers
	box := container.NewVBox()
	scroll := container.NewVScroll(box)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	g.roomBoxes[room] = box
	g.chatScrolls[room] = scroll
    g.listView.Refresh()
}

func (g *GUI) RemoveRoom(room string) {
    for i, r := range g.rooms {
        if r == room {
            g.rooms = append(g.rooms[:i], g.rooms[i+1:]...)
            break
        }
    }
    g.listView.Refresh()
}

func (g *GUI) ShowLobby() {
    rightSide := container.NewBorder(nil, g.bottomBar, nil, nil, g.lobbyScroll)
    split := container.NewHSplit(g.listView, rightSide)
    split.Offset = 0.2
    g.window.SetContent(split)
}

func (g *GUI) UserQuit(msg string) {
    g.quitting = true
    //close the main chat window
    g.window.Close()
    //create a new window just for the quit message
    newWin := fyne.CurrentApp().NewWindow("Session Ended")
    //message label
    label := widget.NewLabel(msg)
    label.Alignment = fyne.TextAlignCenter

    //close button
    closeBtn := widget.NewButton("Close", func() {
        //close window and quit app when clicked
        newWin.Close()
        fyne.CurrentApp().Quit()
    })
    //stack label and button vertically
    vbox := container.NewVBox(
        label,
        closeBtn,
    )
    //center the whole block in the window
    content := container.NewCenter(vbox)
    
    newWin.SetContent(content)
    newWin.Resize(fyne.NewSize(400, 200))
    newWin.Show()
}



// callback will be used to start the actual chat window
func showLoginWindow(a fyne.App, connectCallback func(username string)) fyne.Window {
    loginWin := a.NewWindow("Login")

    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Enter username")

    submit := widget.NewButton("Connect", func() {
        username := usernameEntry.Text
        if username == "" {
            dialog.NewInformation("Error", "Username cannot be empty", loginWin).Show()
            return
        }

        //w.Hide()
        connectCallback(username)
    })

    loginWin.SetContent(container.NewVBox(
        widget.NewLabel("Enter your username:"),
        usernameEntry,
        submit,
    ))

	usernameEntry.OnSubmitted = func(text string) { submit.OnTapped() }

    loginWin.Resize(fyne.NewSize(300, 150))
    loginWin.Show()
	return loginWin
}


func TestWindow() {
    //a := app.New()
    a := app.NewWithID("com.jonny.chatapp")
    loginWin = showLoginWindow(a, func(username string) {
        go func() {
            adapter, msg, err := ConnectToServer(username)
            log.Println(msg)
            if err != nil {
                fyne.Do(func() {
                    fyne.CurrentApp().SendNotification(
                        &fyne.Notification{
                            Title:   "Connection Error",
                            Content: err.Error(),
                        },
                    )
                })
                return
            }
            if adapter == nil {
                fyne.Do(func() {
                    ShowBannedWindow(a, msg)
                    loginWin.Close()
                })
                return
            }

            fyne.Do(func() {
				rooms := getInitRooms(msg)
                gui := MainWindow(a, username, adapter, rooms)
				msg = strings.TrimSuffix(msg, ">")
                dialog.NewInformation("Welcome", msg, gui.window).Show()
                loginWin.Close()
            })
        }()
    })

    a.Run()
}


func ShowBannedWindow(a fyne.App, message string) {
    w := a.NewWindow("Access Denied")
    w.Resize(fyne.NewSize(300, 150))

    w.SetContent(container.NewVBox(
        widget.NewLabel("Access Denied"),
        widget.NewLabel(message),
        widget.NewButton("Close", func() {
            w.Close()
        }),
    ))

    w.Show()
}

func MainWindow(a fyne.App, username string, adapter *ClientAdapter, rooms []string) *GUI{
    mainWin := a.NewWindow("Multi-Room Chat")
    mainWin.Resize(fyne.NewSize(1200, 800))
	//define GUI state
	gui := &GUI{
		roomBoxes: make(map[string]*fyne.Container),
		chatScrolls: make(map[string]*container.Scroll),
		inputChan: make(chan string),
		window: mainWin,
		currentRoom: "",
		rooms: make([]string, 0),
	}
	//create lobby box
	gui.lobbyBox = container.NewVBox()
	gui.lobbyScroll = container.NewVScroll(gui.lobbyBox)
	gui.lobbyScroll.SetMinSize(fyne.NewSize(600, 400))

    // --------------------------
    // BOTTOM: Input bar
    // --------------------------
    input := widget.NewEntry()
    input.SetPlaceHolder("Type a message...")

    //upload file button
    uploadBtn := widget.NewButton("Upload", func() {
        dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
            if err != nil || reader == nil {
                return
            }
            defer reader.Close()

            filePath := reader.URI().Path()

            go func() {
                //uploadedURL, uuid := GenerateImageLink("localhost:8080/uploads", filePath)
                uuid := uuid.New().String()
                url, err := UploadImageToServer("http://localhost:8080", uuid, filePath)
                if err != nil {
                    log.Println("upload error:", err)
                    return
                }
                log.Println("path:", filePath)
                log.Println("url", url)

                // Send the hosted URL as a chat message
                adapter.Outgoing <- "img:" + url
            }()
        }, mainWin)
    })


    sendBtn := widget.NewButton("Send", func() {
        text := input.Text
        if text != "" {
			if gui.currentRoom == "" {
				//send to server
				adapter.Outgoing <- text
				input.SetText("")
			} else {
				gui.roomBoxes[gui.currentRoom].Refresh()
				gui.chatScrolls[gui.currentRoom].ScrollToBottom()
				//send to server
				adapter.Outgoing <- text
				input.SetText("")
			}
        }
    })
    input.OnSubmitted = func(text string) { sendBtn.OnTapped() }

    bottomBar := container.NewBorder(nil, nil, uploadBtn, sendBtn, input)
	gui.bottomBar = bottomBar

    // --------------------------
    // LEFT: Room List
    // --------------------------
    listView := widget.NewList(
        func() int { return len(gui.rooms) },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(id widget.ListItemID, obj fyne.CanvasObject) {
            obj.(*widget.Label).SetText(gui.rooms[id])
        },
    )
	gui.listView = listView
	gui.rooms = rooms
    listView.OnSelected = func(id widget.ListItemID) {
		gui.selectedID = id
		selected := gui.rooms[id]
		//functionality to make clicking on a room function as a join request
		//only send /join if not already in this room
    if gui.currentRoom != selected && selected != "" {
        req := "/join " + selected
        adapter.Outgoing <- req
    }
		
		gui.currentRoom = gui.rooms[id] // <-- set the active room

		var rightSide *fyne.Container
		if gui.currentRoom == "" {
			rightSide = container.NewBorder(nil, bottomBar, nil, nil, gui.lobbyScroll)
			
		} else {
			rightSide = container.NewBorder(nil, bottomBar, nil, nil, gui.chatScrolls[gui.currentRoom])
		}
        split := container.NewHSplit(listView, rightSide)
        split.Offset = 0.2
        mainWin.SetContent(split)
    }

    // --------------------------
    // RIGHT SIDE (default room)
    // --------------------------
	var rightSide *fyne.Container
	if gui.currentRoom == "" {
		rightSide = container.NewBorder(nil, bottomBar, nil, nil, gui.lobbyScroll)
		
	} else {
		rightSide = container.NewBorder(nil, bottomBar, nil, nil, gui.chatScrolls[gui.currentRoom])
	}
	split := container.NewHSplit(listView, rightSide)
    split.Offset = 0.2

    mainWin.SetContent(split)

    // --------------------------
    // RECEIVE PATH: listen for server messages
    // --------------------------
    go func() {
		for msg := range adapter.Incoming {
			log.Println("Incoming message received in GUI loop")
			fyne.Do(func() {
				msg.ExecuteClient(gui)
			})
		}
    }()

	gui.SetRooms(rooms)
    mainWin.Show()
	return gui
}


func getInitRooms(msg string) []string {
	//strip the prefix
	parts := strings.SplitN(msg, ":", 2)
	roomStr := strings.TrimSpace(parts[1])

	//split by spaces
	roomsWithDelim := strings.Fields(roomStr)

	//trim any trailing '>' from room
	var rooms []string
	for _, r := range roomsWithDelim {
		r = strings.TrimSuffix(r, ">")
		if r != "" {
			rooms = append(rooms, r)
		}
	}
	return rooms
}