package socket_test

import (
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/factory"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TwoPlayer", func() {
	var (
		updFactory *factory.Room
		conn1      *testserver.TestPlayerInRoom
		conn2      *testserver.TestPlayerInRoom
	)
	BeforeEach(func(ctx SpecContext) {
		updFactory = factory.NewRoom(protoRoom()).WithLeader(protoPlayer(1).Id)
		conn1, conn2 = createTwoPlayers(ctx)
	})

	It("second player left", func(ctx SpecContext) {
		conn2.Cancel()

		roomMsg := updFactory.WithLobby(conn1.Proto()).Build()

		Expect(conn1.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		Expect(conn2.StartGame(nil)).Should(HaveOccurred())
	})

	It("second player join team", func(ctx SpecContext) {
		const teamName = "team-1"
		By("create team")
		err := conn1.CreateTeam(teamName)
		Expect(err).ShouldNot(HaveOccurred())

		info := conn1.NextMsg(ctx).GetTeamCreated()
		Expect(info).NotTo(BeNil())
		Expect(conn2.NextMsg(ctx).GetTeamCreated()).ShouldNot(BeNil())

		err = conn1.JoinTeam(info.Team.Id)
		Expect(err).ShouldNot(HaveOccurred())

		match := matcher.EqualCmp(
			updFactory.
				Clone().
				WithLobby(conn2.Proto()).
				WithTeams(&gamesvc.Team{
					Id:      info.Team.Id,
					Name:    teamName,
					PlayerA: conn1.Proto(),
				}).
				Build(),
		)
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx)).Should(match)
		}, conn1, conn2)

		By("join team")
		err = conn2.JoinTeam(info.Team.Id)
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory.Clone().WithTeams(
			&gamesvc.Team{
				Id:      info.Team.Id,
				Name:    teamName,
				PlayerA: conn1.Proto(),
				PlayerB: conn2.Proto(),
			},
		).Build()

		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}
	})

	It("transfer leadership once", func(ctx SpecContext) {
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory.
			WithLeader(conn2.ID()).
			WithLobby(conn1.Proto(), conn2.Proto()).
			Build()

		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}
	})

	It("transfer leadership twice", func(ctx SpecContext) {
		By("first transfer")
		err := conn1.TransferLeadership(conn2.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg := updFactory.
			Clone().
			WithLeader(conn2.ID()).
			WithLobby(conn1.Proto(), conn2.Proto()).
			Build()

		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}

		By("second transfer")
		err = conn2.TransferLeadership(conn1.ID())
		Expect(err).ShouldNot(HaveOccurred())

		roomMsg = updFactory.
			WithLeader(conn1.ID()).
			WithLobby(conn1.Proto(), conn2.Proto()).
			Build()
		for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
			Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(roomMsg))
		}
	})

	Describe("in a team", func() {
		const teamName = "our team"

		var updFactory *factory.Room

		BeforeEach(func(ctx SpecContext) {
			teamID := joinSameTeam(ctx, teamName, conn1, conn2)

			updFactory = factory.NewRoom(protoRoom()).
				WithLeader(protoPlayer(1).Id).
				WithTeams(
					&gamesvc.Team{
						Id:      teamID,
						Name:    teamName,
						PlayerA: conn1.Proto(),
						PlayerB: conn2.Proto(),
					},
				)
		})

		It("successfully start game", func(ctx SpecContext) {
			updMsg := updFactory.
				WithStartedGame(true).
				WithPlayerIDTurn(conn1.ID()).
				Build()

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
			updFactory *factory.Room
			turnOrder  []string
		)

		BeforeEach(func(ctx SpecContext) {
			turnOrder = []string{conn1.ID(), conn2.ID()}
			teamID := joinSameTeam(ctx, teamName, conn1, conn2)
			updFactory = factory.NewRoom(protoRoom()).
				WithStartedGame(true).
				WithPlayerIDTurn(turnOrder[0]).
				WithLeader(protoPlayer(1).Id).
				WithTeams(
					&gamesvc.Team{
						Id:      teamID,
						Name:    teamName,
						PlayerA: conn1.Proto(),
						PlayerB: conn2.Proto(),
					},
				)

			err := conn1.StartGame(turnOrder)
			Expect(err).ShouldNot(HaveOccurred())

			matcher := matcher.EqualCmp(updFactory.Build())
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher)
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
				Expect(conn.NextMsg(ctx)).Should(matcher.EqualCmp(updFactory.
					WithPlayerIDTurn(turnOrder[0]).
					WithIsPlaying(true).
					Build(),
				))
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

			matcher := matcher.EqualCmp(updFactory.
				WithPlayerIDTurn(turnOrder[0]).
				WithIsPlaying(true).
				Build(),
			)
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx)).Should(matcher)
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

				matcher := matcher.EqualCmp(updFactory.
					WithPlayerIDTurn(turnOrder[0]).
					WithIsPlaying(true).
					Build(),
				)
				each(func(conn *testserver.TestPlayerInRoom) {
					Expect(conn.NextMsg(ctx)).Should(matcher)
				}, conn1, conn2)
			}, NodeTimeout(time.Second))

			It("right player can end turn", func(ctx SpecContext) {
				err := conn1.EndTurn(1, 2)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(conn2.NextMsg(ctx).GetEndTurn()).Should(matcher.EqualCmp(
					&gamesvc.MsgEndTurn{
						Stats: &gamesvc.Statistics{
							Rights: 1,
							Wrongs: 2,
						},
					},
				))
			}, NodeTimeout(time.Second))

			It("wrong player cannot end turn", func(ctx SpecContext) {
				err := conn2.EndTurn(1, 2)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(conn2.NextMsg(ctx).GetError()).ShouldNot(BeNil())
			}, NodeTimeout(time.Second))
		})
	})
})
