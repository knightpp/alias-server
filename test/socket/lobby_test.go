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

				expectedErr := &game.UnknownMessageTypeError{T: &gamesvc.Message_Start{}}
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

				g.Expect(conn1.Next()).Should(matcher.EqualCmp(
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

func protoRoom() *gamesvc.CreateRoomRequest {
	return &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}
}

func protoPlayer1() *gamesvc.Player {
	return &gamesvc.Player{
		Id:   "id-1",
		Name: "player-1",
	}
}

func protoPlayer2() *gamesvc.Player {
	return &gamesvc.Player{
		Id:   "id-2",
		Name: "player-2",
	}
}

func createTwoPlayers(t *testing.T) (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
	t.Helper()
	g := NewGomegaWithT(t)

	player1 := protoPlayer1()
	player2 := protoPlayer2()
	room := protoRoom()

	srv, err := testserver.CreateAndStart(t)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	p1, err := srv.NewPlayer(ctx, player1)
	g.Expect(err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, room)
	g.Expect(err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	p2, err := srv.NewPlayer(ctx, player2)
	g.Expect(err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	// Sleep to prevent socket messages to pile up that in turn causes
	// occasional out of order packet processing.
	// time.Sleep(50 * time.Millisecond)

	updateRoomReqFactory := updateRoomRequestFactory(room, withLeader(player1.Id))
	roomMsg := updateRoomReqFactory(withLobby(conn1.Proto(), conn2.Proto()))
	for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
		g.Eventually(conn.Poll).Should(matcher.EqualCmp(&gamesvc.Message{
			Message: &gamesvc.Message_UpdateRoom{
				UpdateRoom: roomMsg,
			},
		}))
	}

	return conn1, conn2
}

type updateRoomOption func(*gamesvc.UpdateRoom)

func withLeader(leaderID string) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.LeaderId = leaderID
	}
}

func withLobby(players ...*gamesvc.Player) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.Lobby = players
	}
}

func withTeams(teams ...*gamesvc.Team) updateRoomOption {
	return func(ur *gamesvc.UpdateRoom) {
		ur.Room.Teams = teams
	}
}

func updateRoomRequestFactory(room *gamesvc.CreateRoomRequest, persistentOpts ...updateRoomOption) func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
	return func(opts ...updateRoomOption) *gamesvc.UpdateRoom {
		req := &gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:        testserver.TestUUID,
				Name:      room.Name,
				LeaderId:  "<NOT SET>",
				IsPublic:  room.IsPublic,
				Langugage: room.Langugage,
				Lobby:     nil,
				Teams:     nil,
			},
			Password: nil,
		}

		for _, opt := range persistentOpts {
			opt(req)
		}
		for _, opt := range opts {
			opt(req)
		}

		return req
	}
}
