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

func createTwoPlayers() (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	player1 := protoPlayer1()
	player2 := protoPlayer2()
	room := protoRoom()

	srv, err := testserver.CreateAndStart()
	Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	p1, err := srv.NewPlayer(ctx, player1)
	Expect(err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, room)
	Expect(err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(roomID)
	Expect(err).ShouldNot(HaveOccurred())

	p2, err := srv.NewPlayer(ctx, player2)
	Expect(err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(roomID)
	Expect(err).ShouldNot(HaveOccurred())

	// Sleep to prevent socket messages to pile up that in turn causes
	// occasional out of order packet processing.
	// time.Sleep(50 * time.Millisecond)

	updateRoomReqFactory := updateRoomRequestFactory(room, withLeader(player1.Id))
	roomMsg := updateRoomReqFactory(withLobby(conn1.Proto(), conn2.Proto()))
	for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
		Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
	}

	return conn1, conn2
}

func each(fn func(conn *testserver.TestPlayerInRoom), players ...*testserver.TestPlayerInRoom) {
	for _, conn := range players {
		fn(conn)
	}
}

func createTwoPlayersInATeam(teamName string) (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	conn1, conn2 := createTwoPlayers()

	err := conn1.CreateTeam(teamName)
	Expect(err).ShouldNot(HaveOccurred())

	updMsg, ok := conn1.NextMsg().Message.(*gamesvc.Message_UpdateRoom)
	Expect(ok).To(BeTrue())

	err = conn2.JoinTeam(updMsg.UpdateRoom.Room.Teams[0].Id)
	Expect(err).ShouldNot(HaveOccurred())

	updateRoomReqFactory := updateRoomRequestFactory(protoRoom(), withLeader(conn1.ID()))
	roomMsg := updateRoomReqFactory(withTeams(&gamesvc.Team{
		Id:      "",
		Name:    teamName,
		PlayerA: conn1.Proto(),
		PlayerB: conn2.Proto(),
	}))

	each(func(conn *testserver.TestPlayerInRoom) {
		Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
	}, conn1, conn2)

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

func updateRoomRequestFactory(room *gamesvc.CreateRoomRequest, persistentOpts ...updateRoomOption) func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
	return func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
		req := &gamesvc.UpdateRoom{
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
