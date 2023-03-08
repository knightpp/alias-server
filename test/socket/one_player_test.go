package socket_test

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OnePlayer", func() {
	updFactory := updateRoomRequestFactory(protoRoom(), withLeader(protoPlayer1().Id))

	var conn *testserver.TestPlayerInRoom
	BeforeEach(func(ctx SpecContext) {
		playerProto, room := protoPlayer1(), protoRoom()

		srv, err := testserver.CreateAndStart()
		Expect(err).ShouldNot(HaveOccurred())

		player, err := srv.NewPlayer(ctx, playerProto)
		Expect(err).ShouldNot(HaveOccurred())

		connLocal, err := player.CreateRoomAndJoin(ctx, room)
		Expect(err).ShouldNot(HaveOccurred())

		expectedMsg := updFactory(withLobby(playerProto))
		Expect(connLocal.NextMsg(ctx)).Should(matcher.EqualCmp(expectedMsg))

		conn = connLocal
	})

	It("first message is UpdateRoom", func() {

	})

	It("create team", func(ctx SpecContext) {
		teamName := "my super duper name"
		expectedMsg := updFactory(withTeams(
			&gamesvc.Team{
				Id:      testserver.TestUUID,
				Name:    teamName,
				PlayerA: conn.Proto(),
			},
		))
		err := conn.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(expectedMsg))
	})

	It("start game when no teams", func(ctx SpecContext) {
		err := conn.StartGame([]string{conn.ID()})
		Expect(err).ShouldNot(HaveOccurred())

		expectedErr := game.ErrRoomNoTeams
		Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_Error{
				Error: &gamesvc.MsgError{
					Error: expectedErr.Error(),
				},
			},
		}))
	})

	It("start game when incomplete team", func(ctx SpecContext) {
		err := conn.CreateTeam("super team")
		Expect(err).ShouldNot(HaveOccurred())

		Expect(conn.NextMsg(ctx)).To(matcher.EqualCmp(updFactory(withTeams(&gamesvc.Team{
			Id:      testserver.TestUUID,
			Name:    "super team",
			PlayerA: conn.Proto(),
		}))))

		err = conn.StartGame([]string{conn.ID()})
		Expect(err).ShouldNot(HaveOccurred())

		expectedErr := game.ErrRoomIncompleteTeam
		Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_Error{
				Error: &gamesvc.MsgError{
					Error: expectedErr.Error(),
				},
			},
		}))
	})
})
