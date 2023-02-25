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
	BeforeEach(func() {
		conn1, conn2 = createTwoPlayers()
	})

	It("second player joined should correctly update", func() {

	})

	It("second player left", func() {
		conn2.Cancel()

		roomMsg := updFactory(withLobby(conn1.Proto()))

		Eventually(conn1.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
	})

	It("second player join team", func() {
		const teamName = "team-1"
		By("create team")
		err := conn1.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(conn1.NextMsg()).Should(matcher.EqualCmp(
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
		))

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
			Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	It("transfer leadership once", func() {
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory(
			withLeader(conn2.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	It("transfer leadership twice", func() {
		By("first transfer")
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory(
			withLeader(conn2.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
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
			Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: roomMsg,
				},
			}))
		}
	})

	It("successfully start game", func() {
		return
		err := conn1.StartGame()
		Expect(err).ShouldNot(HaveOccurred())

		updMsg := updFactory(withIsPlaying(true), withPlayerIDTurn(conn1.ID()))

		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
				Message: &gamesvc.Message_UpdateRoom{
					UpdateRoom: updMsg,
				},
			}))
		}
	})
})
