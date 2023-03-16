package socket_test

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/testutil/factory"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OnePlayer", func() {
	var (
		updFactory *factory.Room
		conn       *testserver.TestPlayerInRoom
	)
	BeforeEach(func(ctx SpecContext) {
		updFactory = factory.NewRoom(protoRoom()).WithLeader(protoPlayer(1).Id)

		playerProto, room := protoPlayer(1), protoRoom()

		srv, err := testserver.CreateAndStart()
		Expect(err).ShouldNot(HaveOccurred())

		player, err := srv.NewPlayer(ctx, playerProto)
		Expect(err).ShouldNot(HaveOccurred())

		connLocal, err := player.CreateRoomAndJoin(ctx, room)
		Expect(err).ShouldNot(HaveOccurred())

		expectedMsg := updFactory.Clone().WithLobby(playerProto).Build()
		Expect(connLocal.NextMsg(ctx)).Should(matcher.EqualCmp(expectedMsg))

		conn = connLocal
	})

	It("first message is UpdateRoom", func() {
	})

	It("create team", func(ctx SpecContext) {
		teamName := "my super duper name"

		err := conn.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		resp := conn.NextMsg(ctx).GetTeamCreated()
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.Team.Name, resp.Team.PlayerA, resp.Team.PlayerB).To(Equal(teamName))
	})

	It("start game when no teams", func(ctx SpecContext) {
		err := conn.StartGame(conn.ID())
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
		By("create team")

		err := conn.CreateTeam("super team")
		Expect(err).ShouldNot(HaveOccurred())
		info := conn.NextMsg(ctx).GetTeamCreated()
		Expect(info).ToNot(BeNil())
		msg := updFactory.WithTeams(&gamesvc.Team{
			Id:      info.Team.Id,
			Name:    "super team",
			PlayerA: conn.Proto(),
		}).Build()

		By("join team")

		err = conn.JoinTeam(info.Team.Id)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(msg))

		By("start game")

		err = conn.StartGame(conn.ID())
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
