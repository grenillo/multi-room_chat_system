package common

type Role int
const (
	RoleMember Role = iota
	RoleAdmin
	RoleOwner
	RoleBanned
)

type AdminRequest int
const(
	ReqAll AdminRequest = iota
	ReqMembers
	ReqRooms
	ReqBanned
)

type JoinRequest struct {
	UserName string
	Role Role
	Response chan JoinResponse
}

type JoinResponse struct {
	Status bool
	Message string
	Role interface{}
}

type RoomInfo struct {
	Name string
	Chans RoomChans
}

//channels for each room, currently for joining room or messaging room
type RoomChans struct{
	JoinRoomReq chan string
	MessageRoom chan string
}