package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/test/matcher"
	"github.com/knightpp/alias-server/test/testserver"
	. "github.com/onsi/gomega"
)

func TestJoin_OnePlayer(t *testing.T) {
	g := NewGomegaWithT(t)

	createAndJoin := func(t *testing.T) *testserver.TestPlayerInRoom {
		roomReq := protoRoom()
		playerProto := protoPlayer1()

		srv, err := testserver.CreateAndStart(t)
		g.Expect(err).ShouldNot(HaveOccurred())

		ctx := context.Background()

		player, err := srv.NewPlayer(ctx, playerProto)
		g.Expect(err).ShouldNot(HaveOccurred())

		conn, err := player.CreateRoomAndJoin(ctx, roomReq)
		g.Expect(err).ShouldNot(HaveOccurred())

		var update gamesvc.Message_UpdateRoom
		expectedMsg := &gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:        "",
				Name:      roomReq.Name,
				LeaderId:  playerProto.Id,
				IsPublic:  roomReq.IsPublic,
				Langugage: roomReq.Langugage,
				Lobby:     []*gamesvc.Player{playerProto},
				Teams:     []*gamesvc.Team{},
			},
			Password: nil,
		}

		err = conn.RecvAndAssert(&update)

		update.UpdateRoom.Room.Id = ""
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(update.UpdateRoom).To(matcher.EqualCmp(expectedMsg))

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
				roomReq, playerProto := protoRoom(), protoPlayer1()
				expectedMsg := &gamesvc.UpdateRoom{
					Room: &gamesvc.Room{
						Id:        testserver.TestUUID,
						Name:      roomReq.Name,
						LeaderId:  playerProto.Id,
						IsPublic:  roomReq.IsPublic,
						Langugage: roomReq.Langugage,
						Lobby:     []*gamesvc.Player{},
						Teams: []*gamesvc.Team{
							{
								Id:      testserver.TestUUID,
								Name:    teamName,
								PlayerA: playerProto,
							},
						},
					},
					Password: nil,
				}
				err := p.CreateTeam(teamName)
				g.Expect(err).ShouldNot(HaveOccurred())

				var update gamesvc.Message_UpdateRoom
				err = p.RecvAndAssert(&update)

				g.Expect(err).ShouldNot(HaveOccurred())
				g.Expect(update.UpdateRoom).To(matcher.EqualCmp(expectedMsg))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := createAndJoin(t)

			tt.fn(t, conn)
		})
	}
}

func TestTwoPlayers(t *testing.T) {
	room, player1, player2 := protoRoom(), protoPlayer1(), protoPlayer2()
	tests := []struct {
		name string
		run  func(t *testing.T, conn1, conn2 *testserver.TestPlayerInRoom)
	}{
		{
			name: "second player joined should correctly update",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				g := NewGomegaWithT(t)
				roomMsg := &gamesvc.UpdateRoom{
					Room: &gamesvc.Room{
						Id:        testserver.TestUUID,
						Name:      room.Name,
						LeaderId:  player1.Id,
						IsPublic:  room.IsPublic,
						Langugage: room.Langugage,
						Lobby: []*gamesvc.Player{
							player1,
							player2,
						},
						Teams: []*gamesvc.Team{},
					},
					Password: nil,
				}

				for _, conn := range []*testserver.TestPlayerInRoom{conn1, conn2} {
					g.Eventually(func(g Gomega) *gamesvc.UpdateRoom {
						var update gamesvc.Message_UpdateRoom

						err := conn.RecvAndAssert(&update)

						g.Expect(err).ShouldNot(HaveOccurred())

						return update.UpdateRoom
					}).Should(matcher.EqualCmp(roomMsg))
				}
			},
		},
		{
			name: "second player left",
			run: func(t *testing.T, conn1 *testserver.TestPlayerInRoom, conn2 *testserver.TestPlayerInRoom) {
				conn2.Cancel()

				g := NewGomegaWithT(t)
				roomMsg := &gamesvc.UpdateRoom{
					Room: &gamesvc.Room{
						Id:        testserver.TestUUID,
						Name:      room.Name,
						LeaderId:  player1.Id,
						IsPublic:  room.IsPublic,
						Langugage: room.Langugage,
						Lobby: []*gamesvc.Player{
							player1,
						},
						Teams: []*gamesvc.Team{},
					},
					Password: nil,
				}

				var update gamesvc.Message_UpdateRoom
				err := conn2.RecvAndAssert(&update)
				g.Expect(err).Should(HaveOccurred())

				g.Eventually(func(g Gomega) *gamesvc.UpdateRoom {
					var update gamesvc.Message_UpdateRoom

					err := conn1.RecvAndAssert(&update)

					g.Expect(err).ShouldNot(HaveOccurred())

					return update.UpdateRoom
				}).Should(matcher.EqualCmp(roomMsg))

			},
		},
	}
	for _, tt := range tests {
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
	g := NewGomegaWithT(t)

	srv, err := testserver.CreateAndStart(t)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	p1, err := srv.NewPlayer(ctx, protoPlayer1())
	g.Expect(err).ShouldNot(HaveOccurred())

	roomID, err := p1.CreateRoom(ctx, protoRoom())
	g.Expect(err).ShouldNot(HaveOccurred())

	conn1, err := p1.Join(ctx, roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	p2, err := srv.NewPlayer(ctx, protoPlayer2())
	g.Expect(err).ShouldNot(HaveOccurred())

	conn2, err := p2.Join(ctx, roomID)
	g.Expect(err).ShouldNot(HaveOccurred())

	return conn1, conn2
}
