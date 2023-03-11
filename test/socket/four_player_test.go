package socket_test

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Four players", func() {
	// updFactory := updateRoomRequestFactory(protoRoom(), withLeader(protoPlayer(1).Id))

	var (
	// conn1 *testserver.TestPlayerInRoom
	// conn2 *testserver.TestPlayerInRoom
	// conn3 *testserver.TestPlayerInRoom
	// conn4 *testserver.TestPlayerInRoom
	)
	BeforeEach(func(ctx SpecContext) {
		srv, err := testserver.CreateAndStart()
		Expect(err).ShouldNot(HaveOccurred())

		player1, err := srv.NewPlayer(ctx, protoPlayer(1))
		Expect(err).ShouldNot(HaveOccurred())

		roomID, err := player1.CreateRoom(ctx, protoRoom())
		Expect(err).ShouldNot(HaveOccurred())

		players := srv.CreatePlayers(ctx, 3, func(n int) *gamesvc.Player {
			return protoPlayer(n + 1)
		})

		players = append([]*testserver.TestPlayer{player1}, players...)

		playersInRoom := srv.JoinPlayers(ctx, roomID, players...)

		err = playersInRoom[0].CreateTeam("team 1")
	})
})
