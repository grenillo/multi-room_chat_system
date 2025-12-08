package server

import (
	"multi-room_chat_system/shared"
)

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
			AvailableRooms: getRooms(role),
			ToServer: make(chan shared.MsgMetadata),
			RecvServer: make(chan shared.ExecutableMessage),
			Term: make(chan struct{}),
			Permissions: []string{"/join", "/leave", "/listusers", "/listrooms", "/help", "/quit"},		
		}
	}
}

//function to define admin
func defAdmin(username string, role Role) *Member {
	member := *defMember(username, role)
	member.Permissions = append(member.Permissions, "/kick", "/ban", "/unban","/create", "/delete", "/broadcast")
	return &member
}

//function to define owner
func defOwner(username string, role Role) *Member {
	admin := *defAdmin(username, role)
	admin.Permissions = append(admin.Permissions, "/promote", "/demote", "/shutdown")
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

//function that returns the usage of each member's commands when they enter /help
func getUsage(role Role) []string {
	var usage []string
	if role >= RoleMember {
		member := []string{"/join {room}", "/leave", "/listusers", "/listrooms", "/help", }
		usage = append(usage, member...)
	}
	if role >= RoleAdmin {
		admin := []string{"/kick {user}", "/ban {user}", "/unban {user}","/create {room} {all or staff}", "/delete {room}", "/broadcast {message}"}
		usage = append(usage, admin...)
	}
	if role >= RoleOwner {
		owner := []string{"/promote {user}", "/demote {user}", "/shutdown"}
		usage = append(usage, owner...)
	}
	usage = append(usage, "/quit")
	return usage
}

//dynamically populate user's list of available rooms during runtime
func getRooms(role Role) []string {
	var rooms []string
	//get server state
	s := GetServerState()
	//loop through the current rooms
	for name, room := range s.rooms {
		//if user is at least the role of the room
		if room.permission <= role {
			rooms = append(rooms, name)
		}
	}
	return rooms
}

//function that updates the user's role based on a promote/demote
func (m *Member) updateUserState(role Role, update *UserUpdate) {
	//get available rooms based on role
	newSet := getRooms(role)
	//get set of new rooms
	update.Rooms = getRoomChanges(newSet, m.AvailableRooms)
	//update available rooms
	m.AvailableRooms = newSet
	//set cmds
	var cmds []string
	if role >= RoleMember {
		cmds = append(cmds, "/join", "/leave", "/listusers", "/listrooms", "/help", "/quit")
	}
	if role >= RoleAdmin {
		cmds = append(cmds, "/kick", "/ban", "/unban","/create", "/delete", "/broadcast")
	}
	if role >= RoleOwner {
		cmds = append(cmds, "/promote", "/demote", "/shutdown")
	}
	//update permissions
	m.Permissions = cmds
	m.Role = role
}

//helper function to get the difference in rooms between two slices
func difference(a, b []string) []string {
    //build a lookup map for b
    m := make(map[string]struct{}, len(b))
    for _, x := range b {
        m[x] = struct{}{}
    }

    //collect items from a not in b
    var diff []string
    for _, x := range a {
        if _, found := m[x]; !found {
            diff = append(diff, x)
        }
    }
    return diff
}

//helper function that gets the symmetric difference of rooms between two slices
func getRoomChanges(a, b []string) []string {
    diffAB := difference(a, b) // removed
    diffBA := difference(b, a) // added
    return append(diffAB, diffBA...)
}
