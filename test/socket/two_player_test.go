package socket_test

import (
	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TwoPlayer", func() {
	updFactory := updateRoomRequestFactory(protoRoom(), withLeader(protoPlayer1().Id))

	var (
		conn1 *testserver.TestPlayerInRoom
		conn2 *testserver.TestPlayerInRoom
	)
	BeforeEach(func(ctx SpecContext) {
		conn1, conn2 = createTwoPlayers(ctx)
	})

	It("second player left", func(ctx SpecContext) {
		conn2.Cancel()

		roomMsg := updFactory(withLobby(conn1.Proto()))

		Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
		Expect(conn2.StartGame()).Should(HaveOccurred())
	})

	It("second player join team", func(ctx SpecContext) {
		const teamName = "team-1"
		By("create team")
		err := conn1.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		match := matcher.EqualCmp(
			&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: updFactory(
						withLobby(conn2.Proto()),
						withTeams(&gamesvc.Team{
							Id:      testserver.TestUUID,
							Name:    teamName,
							PlayerA: conn1.Proto(),
						}),
					),
				},
			},
		)
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx)).Should(match)
		}, conn1, conn2)

		By("join team")
		err = conn2.JoinTeam(testserver.TestUUID)
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory(withTeams(
			&gamesvc.Team{
				Id:      testserver.TestUUID,
				Name:    teamName,
				PlayerA: conn1.Proto(),
				PlayerB: conn2.Proto(),
			},
		))

		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	It("transfer leadership once", func(ctx SpecContext) {
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory(
			withLeader(conn2.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	It("transfer leadership twice", func(ctx SpecContext) {
		By("first transfer")
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory(
			withLeader(conn2.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}

		By("second transfer")
		err = conn2.TransferLeadership(conn1.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg = updFactory(
			withLeader(conn1.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	Describe("in a team", func() {
		const teamName = "our team"

		var updFactory func(opts ...updateRoomOption) *gamesvc.UpdateRoom

		JustBeforeEach(func() {
			updFactory = updateRoomRequestFactory(
				protoRoom(),
				withLeader(protoPlayer1().Id),
				withTeams(
					&gamesvc.Team{
						Id:      testserver.TestUUID,
						Name:    teamName,
						PlayerA: conn1.Proto(),
						PlayerB: conn2.Proto(),
					},
				),
			)
		})

		BeforeEach(func(ctx SpecContext) {
			joinSameTeam(ctx, teamName, conn1, conn2)
		})

		It("successfully start game", func(ctx SpecContext) {
			By("start game")
			err := conn1.StartGame()
			Expect(err).ShouldNot(HaveOccurred())

			updMsg := updFactory(
				withIsPlaying(true),
				withPlayerIDTurn(conn1.ID()),
			)

			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_UpdateRoom{
						UpdateRoom: updMsg,
					},
				}))
			}, conn1, conn2)
		})
	})
})
