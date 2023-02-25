package socket_test

import (
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/gomega"
)

func TestTwoPlayers(t *testing.T) {
	t.Parallel()
	updFactory := updateRoomRequestFactory(protoRoom(), withLeader(protoPlayer1().Id))

	tests := []struct {
		name string
		run  func(t *testing.T, conn1, conn2 *testserver.TestPlayerInRoom)
	}{
		{
			name: "second player joined should correctly update",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
			},
		},
		{
			name: "second player left",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)
				conn2.Cancel()

				roomMsg := updFactory(withLobby(conn1.Proto()))

				g.Eventually(conn1.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_UpdateRoom{
						UpdateRoom: roomMsg,
					},
				}))
			},
		},
		{
			name: "second player join team",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)

				const teamName = "team-1"
				err := conn1.CreateTeam(teamName)
				g.Expect(err).ShouldNot(HaveOccurred())

				g.Expect(conn1.NextMsg()).Should(matcher.EqualCmp(
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

				err = conn2.JoinTeam(testserver.TestUUID)
				g.Expect(err).ShouldNot(HaveOccurred())

				roomMsg := updFactory(withTeams(
					&gamesvc.Team{
						Id:      testserver.TestUUID,
						Name:    teamName,
						PlayerA: conn1.Proto(),
						PlayerB: conn2.Proto(),
					},
				))

				for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
					g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
						Message: &gamesvc.Message_UpdateRoom{
							UpdateRoom: roomMsg,
						},
					}))
				}
			},
		},
		{
			name: "transfer leadership once",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)
				err := conn1.TransferLeadership(conn2.ID())
				g.Expect(err).ShouldNot(HaveOccurred())

				roomMsg := updFactory(
					withLeader(conn2.ID()),
					withLobby(conn1.Proto(), conn2.Proto()),
				)
				for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
					g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
						Message: &gamesvc.Message_UpdateRoom{
							UpdateRoom: roomMsg,
						},
					}))
				}
			},
		},
		{
			name: "transfer leadership twice",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)
				err := conn1.TransferLeadership(conn2.ID())
				g.Expect(err).ShouldNot(HaveOccurred())

				roomMsg := updFactory(
					withLeader(conn2.ID()),
					withLobby(conn1.Proto(), conn2.Proto()),
				)
				for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
					g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
						Message: &gamesvc.Message_UpdateRoom{
							UpdateRoom: roomMsg,
						},
					}))
				}

				err = conn2.TransferLeadership(conn1.ID())
				g.Expect(err).ShouldNot(HaveOccurred())

				roomMsg = updFactory(
					withLeader(conn1.ID()),
					withLobby(conn1.Proto(), conn2.Proto()),
				)
				for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
					g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
						Message: &gamesvc.Message_UpdateRoom{
							UpdateRoom: roomMsg,
						},
					}))
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn1, conn2 := createTwoPlayers(t)

			tt.run(t, conn1, conn2)
		})
	}
}
