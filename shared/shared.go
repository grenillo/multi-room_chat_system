package shared

import "encoding/gob"
import "time"

//define all methods a message should have
type ExecutableMessage interface {
	ExecuteServer() //broadcast or command functionality
	ExecuteClient(ui ClientUI)
}

type ClientUI interface {
	Display(room string, text string, broadcast bool)
	ClearRoom(room string)
	ClearLobby()
	SelectRoom(room string)   // highlight/select a room in the list
    DeselectRoom() 
	SetRooms(newRooms []string)
	AddRoom(room string)
	RemoveRoom(room string)
	ShowLobby()
	UserQuit(msg string)
	DisplayImage(room string, url string)
	DisplayJoin(room string, Messages []Message)
}

func Init() {
	gob.Register(&Message{})
	gob.Register(&JoinCmd{})
	gob.Register(&LeaveCmd{})
	gob.Register(&ListUsersCmd{})
	gob.Register(&HelpCmd{})
	gob.Register(&LUResp{})
	gob.Register(&QuitCmd{})
	gob.Register(&KickBanCmd{})
	gob.Register(&CreateCmd{})
	gob.Register(&DeleteCmd{})
	gob.Register(&PromoteDemoteCmd{})
	gob.Register(&BroadcastCmd{})
	gob.Register(&ShutdownCmd{})
	gob.Register(&ListRoomsCmd{})
	gob.Register(&RoomUpdate{})
	gob.Register(&UserUpdate{})
	gob.Register(&UnBanCmd{})
	gob.Register(&GetLog{})
	gob.Register(&UpdateLobby{})
}

type MsgMetadata struct {
	UserName string
	Timestamp time.Time
	Content string
	Flag bool
	Args int
}

type ResponseMD struct {
	Status bool
	ErrMsg string
	CurrentRoom string
}

//message types
type Message struct {
	MsgMetadata
	Response ResponseMD
	Image bool
	URL bool
}

type JoinCmd struct {
	MsgMetadata //inherits metadata
	Room string
	Reply JoinResp
}

type JoinResp struct {
	ResponseMD	//inherits
	Log []Message
}

type LeaveCmd struct {
	MsgMetadata
	Room string
	Reply ResponseMD
	Staff bool
	Log []string
}

type ListUsersCmd struct {
	MsgMetadata
	Reply LUResp
}
type LUResp struct {
	ResponseMD
	Room string
	Users []string
}

type HelpCmd struct {
	MsgMetadata
	Invalid bool
	Reply HelpResp
}
type HelpResp struct {
	ResponseMD
	Usage []string
}

type QuitCmd struct {
	MsgMetadata
	CurrentRoom string
}

type KickBanCmd struct {
	MsgMetadata
	ResponseMD //for displaying error message if not having permission
	Ban bool
	User string
	Sender bool
	InRoom bool
	Msg Message
}

type UnBanCmd struct {
	MsgMetadata
	ResponseMD //for displaying error message if not having permission
	User string
}

type CreateCmd struct {
	MsgMetadata
	ResponseMD //for displaying error message if not having permission
	Room string
	Role int
}

type DeleteCmd struct {
	MsgMetadata
	ResponseMD //for displaying error message if not having permission
	Room string
	InRoom bool
	Log []string
}

type PromoteDemoteCmd struct {
	MsgMetadata
	ResponseMD
	User string
	Promote bool
}

type BroadcastCmd struct {
	MsgMetadata
	ResponseMD
}

type ShutdownCmd struct {
	MsgMetadata
	ResponseMD
	Sender bool
}

type ListRoomsCmd struct {
	MsgMetadata
	LUResp
}

type RoomUpdate struct {
	Create bool
	Room string
}

type UserUpdate struct {
	Promote bool
	Rooms []string
	Log []string
	Current string
}

type GetLog struct {
	Log []string
}

type UpdateLobby struct {
	Update string
}