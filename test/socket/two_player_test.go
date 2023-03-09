package socket_test

import (
	"time"

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

		Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		Expect(conn2.StartGame(nil)).Should(HaveOccurred())
	})

	It("second player join team", func(ctx SpecContext) {
		const teamName = "team-1"
		By("create team")
		err := conn1.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		match := matcher.EqualCmp(
			updFactory(
				withLobby(conn2.Proto()),
				withTeams(&gamesvc.Team{
					Id:      testserver.TestUUID,
					Name:    teamName,
					PlayerA: conn1.Proto(),
				}),
			),
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
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
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
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
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
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}

		By("second transfer")
		err = conn2.TransferLeadership(conn1.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg = updFactory(
			withLeader(conn1.ID()),
			withLobby(conn1.Proto(), conn2.Proto()),
		)
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}
	})

	Describe("in a team", func() {
		const teamName = "our team"

		var updFactory func(opts ...updateRoomOption) *gamesvc.Message

		BeforeEach(func(ctx SpecContext) {
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

			joinSameTeam(ctx, teamName, conn1, conn2)
		})

		It("successfully start game", func(ctx SpecContext) {
			updMsg := updFactory(
				withStartedGame(true),
				withPlayerIDTurn(conn1.ID()),
			)

			err := conn1.StartGame([]string{conn1.ID(), conn2.ID()})

			Expect(err).ShouldNot(HaveOccurred())
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updMsg))
			}, conn1, conn2)
		})
	})

	Context("in a game", func() {
		const teamName = "our team"

		var (
			updFactory func(opts ...updateRoomOption) *gamesvc.Message
			turnOrder  []string
		)

		BeforeEach(func(ctx SpecContext) {
			turnOrder = []string{conn1.ID(), conn2.ID()}
			updFactory = updateRoomRequestFactory(
				protoRoom(),
				withStartedGame(true),
				withPlayerIDTurn(turnOrder[0]),
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

			joinSameTeam(ctx, teamName, conn1, conn2)

			err := conn1.StartGame(turnOrder)
			Expect(err).ShouldNot(HaveOccurred())

			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory()))
			}, conn1, conn2)
		}, NodeTimeout(time.Second))

		It("start turn wrong player", func(ctx SpecContext) {
			err := conn2.StartTurn(time.Second)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(conn2.NextMsg(ctx).GetError()).ToNot(BeNil())
		}, NodeTimeout(time.Second))

		It("start turn right player", func(ctx SpecContext) {
			err := conn1.StartTurn(time.Second)

			Expect(err).ShouldNot(HaveOccurred())
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory(
					withPlayerIDTurn(turnOrder[0]),
					withIsPlaying(true),
				)))
			}, conn1, conn2)
		}, NodeTimeout(time.Second))

		It("start turn with zero duration", func(ctx SpecContext) {
			err := conn1.StartTurn(0)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(conn1.NextMsg(ctx).GetError()).ShouldNot(BeNil())
		}, NodeTimeout(time.Second))

		It("start turn right player twice", func(ctx SpecContext) {
			err := conn1.StartTurn(time.Second)

			By("sending first StartTurn")
			Expect(err).ShouldNot(HaveOccurred())
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory(
					withPlayerIDTurn(turnOrder[0]),
					withIsPlaying(true),
				)))
			}, conn1, conn2)

			By("sending second StartTurn")
			err = conn1.StartTurn(time.Second)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(conn1.NextMsg(ctx).GetError()).ShouldNot(BeNil())
		}, NodeTimeout(time.Second))

		Context("in a turn", func() {
			BeforeEach(func(ctx SpecContext) {
				err := conn1.StartTurn(time.Second)

				Expect(err).ShouldNot(HaveOccurred())
				each(func(conn *testserver.TestPlayerInRoom) {
					Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory(
						withPlayerIDTurn(turnOrder[0]),
						withIsPlaying(true),
					)))
				}, conn1, conn2)
			}, NodeTimeout(time.Second))

			It("end turn should succeed", func(ctx SpecContext) {
				err := conn1.EndTurn(1, 2)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(conn2.NextMsg(ctx).GetEndTurn()).Should(matcher.EqualCmp(
					&gamesvc.MsgEndTurn{
						Rights: 1,
						Wrongs: 2,
					},
				))
			}, NodeTimeout(time.Second))
		})
	})
})
