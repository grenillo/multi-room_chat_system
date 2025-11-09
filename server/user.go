package server

import "multi-room_chat_system/shared"


type AdminReq struct {
	Action AdminRequest
	Reply chan AdminResponse
}

type AdminResponse struct {
	Members []*Member
    Banned  []*User
    Rooms   []*Room
}

type User struct {
	Username string
	Role Role
	Active bool
}

type Member struct {
	User //inherits user
	//current room
	CurrentRoom string
	//channels from the server
	ToServer chan shared.MsgMetadata
	RecvServer chan shared.ExecutableMessage
	//channel to send end to user state
	Term chan struct{}

	//rooms a user can join
	AvailableRooms []string
	//commands the memeber can execute
	Permissions []string
}


//function to define user
func defUser(username string, role Role) *User {
	return &User{
		Username: username,
		Role: role,
	}
}

//function to define member
func defMember(username string, role Role) *Member {
	if role == RoleBanned {
		return &Member {
				User: *defUser(username, role),
		}
	} else {
		return &Member {
			User: *defUser(username, role),
			CurrentRoom: "",
			AvailableRooms: []string{"#general"},
			ToServer: make(chan shared.MsgMetadata),
			RecvServer: make(chan shared.ExecutableMessage),
			Term: make(chan struct{}),
			Permissions: []string{"/join", "/leave", "/listusers", "/help", "/quit"},		
		}
	}
}

//function to define admin
func defAdmin(username string, role Role) *Member {
	member := *defMember(username, role)
	member.AvailableRooms = append(member.AvailableRooms, "#staff")
	member.Permissions = append(member.Permissions, "/kick", "/ban", "/createRoom", "/deleteRoom")
	return &member
}

//function to define owner
func defOwner(username string, role Role) *Member {
	admin := *defAdmin(username, role)
	admin.Permissions = append(admin.Permissions, "/promote", "/demote")
	return &admin
}

//user factory to return the correct user type to the server
func UserFactory(name string, role Role) *Member {
	//return type of member
	if role == RoleMember {
		return defMember(name, role)
	//return type of admin
	} else if role == RoleAdmin {
		return defAdmin(name, role)
	//return type of owner
	} else if role == RoleOwner {
		return defOwner(name, role)
	//return type of user if banned
	} else {
		//override role to indicate banned
		role = RoleBanned
		return defMember(name, role)
	}
}

func getUsage(role Role) []string {
	var usage []string
	if role >= RoleMember {
		usage = append(usage, "/join {room}")
		usage = append(usage, "/leave")
		usage = append(usage, "/listusers")
	}
	if role >= RoleAdmin {
		usage = append(usage, "/kick {user}")
		usage = append(usage, "/ban {user}")
		usage = append(usage, "/createRoom {roomName} {rolePermission}")
		usage = append(usage, "/deleteRoom {roomName}")
	}
	if role >= RoleOwner {
		usage = append(usage, "/promote {user}")
		usage = append(usage, "/demote {user}")
	}
	usage = append(usage, "/quit")
	return usage
}