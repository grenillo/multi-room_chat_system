package server

import "multi-room_chat_system/common"


type AdminRequest struct {
	Action common.AdminRequest
	Reply chan AdminResponse
}

type AdminResponse struct {
	Members []*Member
    Banned  []*User
    Rooms   []*Room
}

type User struct {
	Username string
	Role common.Role
}

type Member struct {
	User //inherits user
	//rooms a user can join
	AvailableRooms []common.RoomInfo
	//commands the memeber can execute
	Permissions []string
}

type Admin struct {
	Member //inherits member
	//channel to request from the server as an admin
	adminReq chan AdminRequest

}

type Owner struct {
	Admin	//inherits admin
	//chanel to get all admins
	reqAdmin chan []*Admin
}

//function to define user
func defUser(username string, role common.Role) *User {
	return &User{
		Username: username,
		Role: role,
	}
}

//function to define member
func defMember(username string, role common.Role) *Member {
	return &Member {
		User: *defUser(username, role),
		AvailableRooms: []common.RoomInfo{},
		Permissions: []string{},		
	}
}

//function to define admin
func defAdmin(username string, role common.Role) *Admin {
	return &Admin{
		Member: *defMember(username, role),
		adminReq: make(chan AdminRequest),
	}
}

//function to define owner
func defOwner(username string, role common.Role) *Owner {
	return &Owner{
		Admin: *defAdmin(username, role),
		reqAdmin: make(chan []*Admin),
	}
}

//user factory to return the correct user type to the server
func UserFactory(name string, role common.Role) interface{} {
	//return type of member
	if role == common.RoleMember {
		return defMember(name, role)
	//return type of admin
	} else if role == common.RoleAdmin {
		return defAdmin(name, role)
	//return type of owner
	} else if role == common.RoleOwner {
		return defOwner(name, role)
	//return type of user if banned
	} else {
		//override role to indicate banned
		role = common.RoleBanned
		return defUser(name, role)
	}
}