package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/gomega"
)

func protoRoom() *gamesvc.CreateRoomRequest {
	return &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}
}

func protoPlayer1() *gamesvc.Player {
	return &gamesvc.Player{
		Id:   "id-1",
		Name: "player-1",
	}
}

func protoPlayer2() *gamesvc.Player {
	return &gamesvc.Player{
		Id:   "id-2",
		Name: "player-2",
	}
}

func createTwoPlayers(t *testing.T) (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	t.Helper()
	g := NewGomegaWithT(t)

	player1 := protoPlayer1()
	player2 := protoPlayer2()
	room := protoRoom()

	srv, err := testserver.CreateAndStart(t)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	p1, err := srv.NewPlayer(ctx, player1)
	g.Expect(err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, room)
	g.Expect(err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	p2, err := srv.NewPlayer(ctx, player2)
	g.Expect(err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	// Sleep to prevent socket messages to pile up that in turn causes
	// occasional out of order packet processing.
	// time.Sleep(50 * time.Millisecond)

	updateRoomReqFactory := updateRoomRequestFactory(room, withLeader(player1.Id))
	roomMsg := updateRoomReqFactory(withLobby(conn1.Proto(), conn2.Proto()))
	for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
		g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
	}

	return conn1, conn2
}

type updateRoomOption func(*gamesvc.UpdateRoom)

func withLeader(leaderID string) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.LeaderId = leaderID
	}
}

func withLobby(players ...*gamesvc.Player) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.Lobby = players
	}
}

func withTeams(teams ...*gamesvc.Team) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.Teams = teams
	}
}

func updateRoomRequestFactory(room *gamesvc.CreateRoomRequest, persistentOpts ...updateRoomOption) func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
	return func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
		req := &gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:        testserver.TestUUID,
				Name:      room.Name,
				LeaderId:  "",
				IsPublic:  room.IsPublic,
				Langugage: room.Langugage,
				Lobby:     nil,
				Teams:     nil,
			},
			Password: nil,
		}

		for _, opt := range persistentOpts {
			opt(req)
		}
		for _, opt := range opts {
			opt(req)
		}

		if req.Room.LeaderId == "" {
			panic("LeaderId must not be empty")
		}

		return req
	}
}
