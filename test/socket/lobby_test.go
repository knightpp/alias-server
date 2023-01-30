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

	playerProto := &gamesvc.Player{
		Id:   "id-1",
		Name: "player-1",
	}
	roomReq := &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}

	createAndJoin := func(t *testing.T) *testserver.TestPlayerInRoom {
		// FIXME: clone
		roomReq := *roomReq
		playerProto := *playerProto

		srv, err := testserver.CreateAndStart(t)
		g.Expect(err).ShouldNot(HaveOccurred())

		ctx := context.Background()

		player, err := srv.NewPlayer(ctx, &playerProto)
		g.Expect(err).ShouldNot(HaveOccurred())

		conn, err := player.CreateRoomAndJoin(ctx, &roomReq)
		g.Expect(err).ShouldNot(HaveOccurred())

		var update gamesvc.Message_UpdateRoom
		expectedMsg := &gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:        "",
				Name:      roomReq.Name,
				LeaderId:  playerProto.Id,
				IsPublic:  roomReq.IsPublic,
				Langugage: roomReq.Langugage,
				Lobby:     []*gamesvc.Player{&playerProto},
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

func TestJoin_SecondPlayerJoined(t *testing.T) {
	g := NewGomegaWithT(t)

	srv, err := testserver.CreateAndStart(t)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	room := &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}
	player1 := &gamesvc.Player{
		Id:   "id-1",
		Name: "player-1",
	}
	player2 := &gamesvc.Player{
		Id:   "id-2",
		Name: "player-2",
	}

	conn1, conn2 := func() (*testserver.TestPlayerInRoom, *testserver.TestPlayerInRoom) {
		p1, err := srv.NewPlayer(ctx, player1)
		g.Expect(err).ShouldNot(HaveOccurred())

		roomID, err := p1.CreateRoom(ctx, room)
		g.Expect(err).ShouldNot(HaveOccurred())

		conn1, err := p1.Join(ctx, roomID)
		g.Expect(err).ShouldNot(HaveOccurred())

		p2, err := srv.NewPlayer(ctx, player2)
		g.Expect(err).ShouldNot(HaveOccurred())

		conn2, err := p2.Join(ctx, roomID)
		g.Expect(err).ShouldNot(HaveOccurred())

		return conn1, conn2
	}()

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

			err = conn.RecvAndAssert(&update)

			g.Expect(err).ShouldNot(HaveOccurred())

			return update.UpdateRoom
		}).WithContext(ctx).Should(matcher.EqualCmp(roomMsg))
	}
}
