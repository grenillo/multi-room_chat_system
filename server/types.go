package server

import "multi-room_chat_system/shared"

type Role int
const (
	RoleBanned Role = iota
	RoleMember
	RoleAdmin
	RoleOwner
)

type AdminRequest int
const(
	ReqAll AdminRequest = iota
	ReqMembers
	ReqRooms
	ReqBanned
)


//SERVER TYPES
type ServerJoinRequest string //join the server (client -> server)

type ServerJoinResponse struct { //server response to join request (server -> client)
	Status bool
	Message string
	Role *Member
}

type JoinRoomReq struct { //client request to join a room (client -> server) or (server -> room)
	Room string //room name
	Username string
	Resp chan JoinRoomResp //RPC listening on this channel (is only included bc passed from server to room)
}

//ROOM TYPES
type JoinRoomResp struct { //room response to join request (room -> client)
	Status bool
	History []string
	ToUser chan shared.ExecutableMessage
}

