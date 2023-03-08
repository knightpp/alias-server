package socket_test

import (
	"context"

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

func createTwoPlayers(ctx context.Context) (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	player1 := protoPlayer1()
	player2 := protoPlayer2()
	room := protoRoom()
	updateRoomReqFactory := updateRoomRequestFactory(room, withLeader(player1.Id))

	srv, err := testserver.CreateAndStart()
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	p1, err := srv.NewPlayer(ctx, player1)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, room)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(roomID)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	ExpectWithOffset(1, conn1.NextMsg(ctx)).Should(matcher.EqualCmp(updateRoomReqFactory(withLobby(player1))))

	p2, err := srv.NewPlayer(ctx, player2)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(roomID)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())

	match := matcher.EqualCmp(updateRoomReqFactory(withLobby(player1, player2)))
	for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
		ExpectWithOffset(1, conn.NextMsg(ctx)).Should(match)
	}

	return conn1, conn2
}

func each(fn func(conn *testserver.TestPlayerInRoom), players ...*testserver.TestPlayerInRoom) {
	for _, conn := range players {
		fn(conn)
	}
}

func joinSameTeam(
	ctx context.Context,
	teamName string,
	conn1, conn2 *testserver.TestPlayerInRoom,
) {
	err := conn1.CreateTeam(teamName)
	Expect(err).ShouldNot(HaveOccurred())

	updMsg, ok := conn1.NextMsg(ctx).Message.(*gamesvc.Message_UpdateRoom)
	Expect(ok).To(BeTrue())

	teamID := updMsg.UpdateRoom.Room.Teams[0].Id

	err = conn2.JoinTeam(teamID)
	Expect(err).ShouldNot(HaveOccurred())

	updateRoomReqFactory := updateRoomRequestFactory(protoRoom(), withLeader(conn1.ID()))
	roomMsg := updateRoomReqFactory(withTeams(&gamesvc.Team{
		Id:      teamID,
		Name:    teamName,
		PlayerA: conn1.Proto(),
		PlayerB: conn2.Proto(),
	}))

	each(func(conn *testserver.TestPlayerInRoom) {
		Eventually(conn.Poll).Should(matcher.EqualCmp(roomMsg))
	}, conn1, conn2)
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

func withStartedGame(started bool) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.IsGameStarted = started
	}
}

func withIsPlaying(playing bool) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.IsPlaying = playing
	}
}

func withPlayerIDTurn(id string) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.PlayerIdTurn = id
	}
}

func updateRoomRequestFactory(
	room *gamesvc.CreateRoomRequest,
	persistentOpts ...updateRoomOption,
) func(opts ...updateRoomOption) *gamesvc.Message {
	return func(opts ...updateRoomOption) *gamesvc.Message {
		msg := &gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:           testserver.TestUUID,
				Name:         room.Name,
				LeaderId:     "",
				IsPublic:     room.IsPublic,
				Langugage:    room.Langugage,
				Lobby:        nil,
				Teams:        nil,
				IsPlaying:    false,
				PlayerIdTurn: "",
			},
			Password: nil,
		}

		for _, opt := range persistentOpts {
			opt(msg)
		}
		for _, opt := range opts {
			opt(msg)
		}

		if msg.Room.LeaderId == "" {
			panic("LeaderId must not be empty")
		}

		return &gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: msg,
			},
		}
	}
}
