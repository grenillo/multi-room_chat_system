package shared

import "encoding/gob"
import "time"

//define all methods a message should have
type ExecutableMessage interface {
	ExecuteServer() //broadcast or command functionality
	ExecuteClient()
}

func Init() {
	gob.Register(&Message{})
	gob.Register(&JoinCmd{})
	gob.Register(&LeaveCmd{})
	gob.Register(&ListUsersCmd{})
	gob.Register(&HelpCmd{})
	gob.Register(&LUResp{})
}

type MsgMetadata struct {
	UserName string
	Timestamp time.Time
	Content string
	Flag bool
}

type ResponseMD struct {
	Status bool
	ErrMsg string
}

//message types
type Message struct {
	MsgMetadata
	Response ResponseMD
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