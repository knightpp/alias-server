package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/game"
	"github.com/knightpp/alias-server/internal/testutil/matcher"
	"github.com/knightpp/alias-server/internal/testutil/testserver"
	. "github.com/onsi/gomega"
)

func TestJoin_OnePlayer(t *testing.T) {
	t.Parallel()
	updFactory := updateRoomRequestFactory(protoRoom(), withLeader(protoPlayer1().Id))

	createAndJoin := func(t *testing.T) *testserver.TestPlayerInRoom {
		g := NewGomegaWithT(t)
		playerProto, room := protoPlayer1(), protoRoom()

		srv, err := testserver.CreateAndStart(t)
		g.Expect(err).ShouldNot(HaveOccurred())

		ctx := context.Background()

		player, err := srv.NewPlayer(ctx, playerProto)
		g.Expect(err).ShouldNot(HaveOccurred())

		conn, err := player.CreateRoomAndJoin(ctx, room)
		g.Expect(err).ShouldNot(HaveOccurred())

		expectedMsg := updFactory(withLobby(playerProto))
		g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: expectedMsg,
			},
		}))

		return conn
	}

	tests := []struct {
		name string
		fn   func(t *testing.T, p *testserver.TestPlayerInRoom)
	}{
		{
			name: "first message is UpdateRoom",
			fn: func(t *testing.T, p *testserver.TestPlayerInRoom) {
			},
		},
		{
			name: "create team",
			fn: func(t *testing.T, p *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)
				teamName := "my super duper name"
				expectedMsg := updFactory(withTeams(
					&gamesvc.Team{
						Id:      testserver.TestUUID,
						Name:    teamName,
						PlayerA: p.Proto(),
					},
				))
				err := p.CreateTeam(teamName)
				g.Expect(err).ShouldNot(HaveOccurred())

				g.Eventually(p.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_UpdateRoom{
						UpdateRoom: expectedMsg,
					},
				}))
			},
		},
		{
			name: "start game when no teams",
			fn: func(t *testing.T, p *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)

				err := p.StartGame()
				g.Expect(err).ShouldNot(HaveOccurred())

				expectedErr := game.ErrStartNoTeams
				g.Eventually(p.PollRaw).Should(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_Error{
						Error: &gamesvc.MsgError{
							Error: expectedErr.Error(),
						},
					},
				}))
			},
		},
		{
			name: "start game when incomplete team",
			fn: func(t *testing.T, p *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)

				err := p.CreateTeam("super team")
				g.Expect(err).ShouldNot(HaveOccurred())

				g.Expect(p.NextMsg()).To(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_UpdateRoom{
						UpdateRoom: updFactory(withTeams(&gamesvc.Team{
							Id:      testserver.TestUUID,
							Name:    "super team",
							PlayerA: p.Proto(),
						})),
					},
				}))

				err = p.StartGame()
				g.Expect(err).ShouldNot(HaveOccurred())

				expectedErr := game.ErrStartIncompleteTeam
				g.Eventually(p.PollRaw).Should(matcher.EqualCmp(&gamesvc.Message{
					Message: &gamesvc.Message_Error{
						Error: &gamesvc.MsgError{
							Error: expectedErr.Error(),
						},
					},
				}))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn := createAndJoin(t)

			tt.fn(t, conn)
		})
	}
}
