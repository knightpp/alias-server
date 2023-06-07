package socket_test

import (
	"strconv"

	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
)

func protoRoom() *gamesvc.CreateRoomRequest {
	return &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}
}

func protoPlayer(n int) *gamesvc.Player {
	nStr := strconv.FormatInt(int64(n), 10)
	return &gamesvc.Player{
		Id:   "id-" + nStr,
		Name: "player-" + nStr,
	}
}
