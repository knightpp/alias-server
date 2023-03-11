package socket_test

import (
	"context"
	"strconv"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/factory"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
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

func protoPlayer(n int) *gamesvc.Player {
	nStr := strconv.FormatInt(int64(n), 10)
	return &gamesvc.Player{
		Id:   "id-" + nStr,
		Name: "player-" + nStr,
	}
}

func createTwoPlayers(ctx context.Context) (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	player1 := protoPlayer(1)
	player2 := protoPlayer(2)
	room := protoRoom()
	updFactory := factory.NewRoom(room).WithLeader(player1.Id)

	srv, err := testserver.CreateAndStart()
	Expect(err).ShouldNot(HaveOccurred())

	p1, err := srv.NewPlayer(ctx, player1)
	Expect(err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, room)
	Expect(err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(roomID)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory.WithLobby(player1).Build()))

	p2, err := srv.NewPlayer(ctx, player2)
	Expect(err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(roomID)
	Expect(err).ShouldNot(HaveOccurred())

	match := matcher.EqualCmp(updFactory.WithLobby(player1, player2).Build())
	for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
		Expect(conn.NextMsg(ctx)).Should(match)
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
) (teamID string) {
	By("create team")
	err := conn1.CreateTeam(teamName)
	Expect(err).ShouldNot(HaveOccurred())

	teamInfo := conn1.NextMsg(ctx).GetTeamCreated()
	Expect(teamInfo).ShouldNot(BeNil())
	Expect(conn2.NextMsg(ctx).GetTeamCreated()).ShouldNot(BeNil())

	teamID = teamInfo.Team.Id

	By("join team")
	err = conn1.JoinTeam(teamID)
	Expect(err).ShouldNot(HaveOccurred())

	updFactory := factory.NewRoom(protoRoom()).WithLeader(conn1.ID())
	team := &gamesvc.Team{
		Id:      teamID,
		Name:    teamName,
		PlayerA: conn1.Proto(),
		PlayerB: nil,
	}
	roomMsg := updFactory.Clone().WithTeams(team).WithLobby(conn2.Proto()).Build()

	Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
	Expect(conn2.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))

	err = conn2.JoinTeam(teamID)
	Expect(err).ShouldNot(HaveOccurred())

	team.PlayerB = conn2.Proto()
	roomMsg = updFactory.WithTeams(team).Build()

	Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
	Expect(conn2.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
	return teamID
}
