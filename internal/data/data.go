package data

type (
	RoomID   string
	TeamID   string
	PlayerID string
)

type Room struct {
	ID       RoomID
	Name     string
	Leader   PlayerID
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
	ID          PlayerID
	Name        string
	GravatarURL string
}

type CreateRoomRequest struct {
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

type UserSimpleLoginRequest struct {
	Name  string
	Email *string
}

type UserSimpleLoginResponse struct {
	Player Player
}

type JoinRoomRequest struct {
	RoomID   RoomID
	PlayerID PlayerID
}
