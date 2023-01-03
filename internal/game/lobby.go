package game

type Lobby struct {
	Players []*Player
}

func NewLobby() *Lobby {
	return &Lobby{}
}
