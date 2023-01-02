package game

type Team struct {
	PlayerA *Player
	PlayerB *Player
}

func NewTeam(playerA, playerB *Player) Team {
	return Team{
		PlayerA: playerA,
		PlayerB: playerB,
	}
}
