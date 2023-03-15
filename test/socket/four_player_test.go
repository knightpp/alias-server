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

var _ = Describe("Four players", func() {
	var (
		conn1        *testserver.TestPlayerInRoom
		conn2        *testserver.TestPlayerInRoom
		conn3        *testserver.TestPlayerInRoom
		conn4        *testserver.TestPlayerInRoom
		firstTeamId  string
		secondTeamId string
		updFactory   *factory.Room
	)
	BeforeEach(func(ctx SpecContext) {
		const (
			firstTeamName  = "team 1"
			secondTeamName = "team 2"
		)

		By("create test server")
		srv, err := testserver.CreateAndStart()
		Expect(err).ShouldNot(HaveOccurred())

		By("create fist player")
		player1, err := srv.NewPlayer(ctx, protoPlayer(1))
		Expect(err).ShouldNot(HaveOccurred())

		By("create room")
		roomID, err := player1.CreateRoom(ctx, protoRoom())
		Expect(err).ShouldNot(HaveOccurred())

		By("create other players")
		players := srv.CreatePlayers(ctx, 3, func(n int) *gamesvc.Player {
			return protoPlayer(n + 1)
		})

		players = append([]*testserver.TestPlayer{player1}, players...)

		By("join players")
		playersInRoom := srv.JoinPlayers(ctx, roomID, players...)

		By("create first team")
		err = playersInRoom[0].CreateTeam(firstTeamName)
		Expect(err).ShouldNot(HaveOccurred())

		firstTeamInfo := playersInRoom[0].NextMsg(ctx).GetTeamCreated()
		Expect(firstTeamInfo).NotTo(BeNil())

		firstTeamId = firstTeamInfo.Team.Id

		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetTeamCreated()).ShouldNot(BeNil())
		}, playersInRoom[1:]...)

		By("create second team")
		err = playersInRoom[2].CreateTeam(secondTeamName)
		Expect(err).ShouldNot(HaveOccurred())

		secondTeamInfo := playersInRoom[0].NextMsg(ctx).GetTeamCreated()
		Expect(secondTeamInfo).NotTo(BeNil())

		secondTeamId = secondTeamInfo.Team.Id

		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetTeamCreated()).ShouldNot(BeNil())
		}, playersInRoom[1:]...)

		By("players join teams")
		err = playersInRoom[0].JoinTeam(firstTeamInfo.Team.Id)
		Expect(err).Should(BeNil())
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetUpdateRoom()).ShouldNot(BeNil())
		}, playersInRoom...)

		err = playersInRoom[1].JoinTeam(firstTeamInfo.Team.Id)
		Expect(err).Should(BeNil())
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetUpdateRoom()).ShouldNot(BeNil())
		}, playersInRoom...)

		err = playersInRoom[2].JoinTeam(secondTeamInfo.Team.Id)
		Expect(err).Should(BeNil())
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetUpdateRoom()).ShouldNot(BeNil())
		}, playersInRoom...)

		err = playersInRoom[3].JoinTeam(secondTeamInfo.Team.Id)
		Expect(err).Should(BeNil())
		each(func(conn *testserver.TestPlayerInRoom) {
			Expect(conn.NextMsg(ctx).GetUpdateRoom()).ShouldNot(BeNil())
		}, playersInRoom...)

		conn1 = playersInRoom[0]
		conn2 = playersInRoom[1]
		conn3 = playersInRoom[2]
		conn4 = playersInRoom[3]

		updFactory = factory.NewRoom(protoRoom()).
			WithLeader(conn1.ID()).
			WithTeams(
				&gamesvc.Team{
					Id:      firstTeamId,
					Name:    firstTeamName,
					PlayerA: conn1.Proto(),
					PlayerB: conn2.Proto(),
				},
				&gamesvc.Team{
					Id:      secondTeamId,
					Name:    secondTeamName,
					PlayerA: conn3.Proto(),
					PlayerB: conn4.Proto(),
				},
			)
	}, NodeTimeout(time.Second))

	Context("in game", func() {
		BeforeEach(func(ctx SpecContext) {
			updFactory = updFactory.WithPlayerIDTurn(conn1.ID())

			turnOrder := []string{
				conn1.ID(), conn2.ID(), conn3.ID(), conn4.ID(),
			}
			By("start game")
			err := conn1.StartGame(turnOrder)
			Expect(err).ShouldNot(HaveOccurred())
			match := matcher.EqualCmp(&gamesvc.MsgStartGame{
				Turns: turnOrder,
			})
			each(func(conn *testserver.TestPlayerInRoom) {
				Expect(conn.NextMsg(ctx).GetStartGame()).Should(match)
			}, conn1, conn2, conn3, conn4)
		}, NodeTimeout(time.Second))

		It("succeed", func() {

		})

		It("when game is not started sending word should error", func(ctx SpecContext) {
			err := conn1.Word("word")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(conn1.NextMsgUnpack(ctx)).Should(BeAssignableToTypeOf(&gamesvc.MsgError{}))
		}, NodeTimeout(time.Second))

		Context("in turn", func() {
			BeforeEach(func(ctx SpecContext) {
				err := conn1.StartTurn(time.Minute)
				Expect(err).ShouldNot(HaveOccurred())

				each(func(conn *testserver.TestPlayerInRoom) {
					Expect(conn.NextMsg(ctx).GetStartTurn()).Should(matcher.EqualCmp(&gamesvc.MsgStartTurn{
						DurationMs: uint64(time.Minute.Milliseconds()),
					}))
				}, conn1, conn2, conn3, conn4)
			}, NodeTimeout(time.Second))

			When("wrong player", func() {
				It("sends word", func(ctx SpecContext) {
					err := conn4.Word("abc")

					Expect(err).ShouldNot(HaveOccurred())
					Expect(conn4.NextMsg(ctx).GetError()).ShouldNot(BeNil())
				}, NodeTimeout(time.Second))

				It("sends end turn", func(ctx SpecContext) {
					err := conn3.EndTurn(0, 20)

					Expect(err).ShouldNot(HaveOccurred())
					Expect(conn3.NextMsg(ctx).GetError()).ShouldNot(BeNil())
				}, NodeTimeout(time.Second))

				It("sends end game", func(ctx SpecContext) {
					each(func(conn *testserver.TestPlayerInRoom) {
						err := conn.EndGame()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(conn.NextMsg(ctx).GetError()).ShouldNot(BeNil())
					}, conn2, conn3, conn4)
				})
			})

			When("right player", func() {
				It("sends word", func(ctx SpecContext) {
					err := conn1.Word("abc")

					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgWord{
							Word: "abc",
						}))
					}, conn3, conn4)
				}, NodeTimeout(time.Second))

				It("sends end turn", func(ctx SpecContext) {
					err := conn1.EndTurn(10, 15)

					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgEndTurn{
							Stats: &gamesvc.Statistics{
								Rights: 10,
								Wrongs: 15,
							},
						}))
					}, conn2, conn3, conn4)
				}, NodeTimeout(time.Second))

				It("end game with non zero statistics", func(ctx SpecContext) {
					By("send word")
					err := conn1.Word("abc")
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgWord{
							Word: "abc",
						}))
					}, conn3, conn4)

					By("end turn")
					err = conn1.EndTurn(10, 15)
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgEndTurn{
							Stats: &gamesvc.Statistics{
								Rights: 10,
								Wrongs: 15,
							},
						}))
					}, conn2, conn3, conn4)

					By("end game")
					err = conn1.EndGame()
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsg(ctx).GetResults()).Should(matcher.EqualCmp(&gamesvc.MsgResults{
							TeamIdToStats: map[string]*gamesvc.Statistics{
								firstTeamId: {
									Rights: 10,
									Wrongs: 15,
								},
							},
						}))
					}, conn1, conn2, conn3, conn4)
				}, NodeTimeout(time.Second))

				It("end turn end game start game", func(ctx SpecContext) {
					By("end turn")
					err := conn1.EndTurn(10, 10)
					Expect(err).Should(BeNil())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgEndTurn{
							Stats: &gamesvc.Statistics{
								Rights: 10,
								Wrongs: 10,
							},
						}))
					}, conn2, conn3, conn4)

					By("end game")
					err = conn1.EndGame()
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsg(ctx).GetResults()).Should(matcher.EqualCmp(&gamesvc.MsgResults{
							TeamIdToStats: map[string]*gamesvc.Statistics{
								firstTeamId: {
									Rights: 10,
									Wrongs: 10,
								},
							},
						}))
					}, conn1, conn2, conn3, conn4)

					By("start game")
					turnOrder := []string{
						conn1.ID(), conn2.ID(), conn3.ID(), conn4.ID(),
					}
					err = conn1.StartGame(turnOrder)
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgStartGame{
							Turns: turnOrder,
						}))
					}, conn1, conn2, conn3, conn4)

					By("start turn")
					err = conn1.StartTurn(time.Minute)
					Expect(err).ShouldNot(HaveOccurred())
					each(func(conn *testserver.TestPlayerInRoom) {
						Expect(conn.NextMsgUnpack(ctx)).Should(matcher.EqualCmp(&gamesvc.MsgStartTurn{
							DurationMs: uint64(time.Minute.Milliseconds()),
						}))
					}, conn1, conn2, conn3, conn4)
				}, NodeTimeout(time.Second))
			})
		})

		When("right player", func() {
			It("sends end game", func(ctx SpecContext) {
				err := conn1.EndGame()

				Expect(err).ShouldNot(HaveOccurred())
				each(func(conn *testserver.TestPlayerInRoom) {
					Expect(conn.NextMsg(ctx).GetResults()).Should(matcher.EqualCmp(&gamesvc.MsgResults{
						TeamIdToStats: nil,
					}))
				}, conn1, conn2, conn3, conn4)
			}, NodeTimeout(time.Second))
		})
	})
})
