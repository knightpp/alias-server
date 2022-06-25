package data

type (
	RoomID   []byte
	TeamID   []byte
	PlayerID []byte
)

type Room struct {
	ID       RoomID
	Name     string
	IsPublic bool
	Language string
	Lobby    []Player
	Teams    []Team
}

type Team struct {
	ID      TeamID
	Name    string
	PlayerA Player
	PlayerB Player
}

type Player struct {
	ID   PlayerID
	Name string
}

type CreateRoomOptions struct {
	Name     string
	IsPublic bool
	Language string
	Password *string
}

type CreateRoomResponse struct {
	ID RoomID
}

type ListRoomsResponse struct {
	Rooms []Room
}
