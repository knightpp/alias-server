package socket_test

import (
	"context"

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
	BeforeEach(func() {
		playerProto, room := protoPlayer1(), protoRoom()

		srv, err := testserver.CreateAndStart()
		Expect(err).ShouldNot(HaveOccurred())

		ctx := context.Background()

		player, err := srv.NewPlayer(ctx, playerProto)
		Expect(err).ShouldNot(HaveOccurred())

		connLocal, err := player.CreateRoomAndJoin(ctx, room)
		Expect(err).ShouldNot(HaveOccurred())

		expectedMsg := updFactory(withLobby(playerProto))
		Eventually(connLocal.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: expectedMsg,
			},
		}))

		conn = connLocal
	})

	It("first message is UpdateRoom", func() {

	})

	It("create team", func() {
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

		Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: expectedMsg,
			},
		}))
	})

	It("start game when no teams", func() {
		err := conn.StartGame()
		Expect(err).ShouldNot(HaveOccurred())

		expectedErr := game.ErrStartNoTeams
		Eventually(conn.PollRaw).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_Error{
				Error: &gamesvc.MsgError{
					Error: expectedErr.Error(),
				},
			},
		}))
	})

	It("start game when incomplete team", func() {
		err := conn.CreateTeam("super team")
		Expect(err).ShouldNot(HaveOccurred())

		Expect(conn.NextMsg()).To(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: updFactory(withTeams(&gamesvc.Team{
					Id:      testserver.TestUUID,
					Name:    "super team",
					PlayerA: conn.Proto(),
				})),
			},
		}))

		err = conn.StartGame()
		Expect(err).ShouldNot(HaveOccurred())

		expectedErr := game.ErrStartIncompleteTeam
		Eventually(conn.PollRaw).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_Error{
				Error: &gamesvc.MsgError{
					Error: expectedErr.Error(),
				},
			},
		}))
	})
})
