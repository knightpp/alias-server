package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/test/matcher"
	"github.com/knightpp/alias-server/test/testserver"
	. "github.com/onsi/gomega"
)

func TestJoin_FirstMessageIsUpdateRoom(t *testing.T) {
	g := NewGomegaWithT(t)

	srv, err := testserver.CreateAndStart(t)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	player, err := srv.NewPlayer(ctx, &gamesvc.Player{
		Id:   "id-1",
		Name: "player-1",
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	room := &gamesvc.CreateRoomRequest{
		Name:      "room-1",
		IsPublic:  true,
		Langugage: "UA",
		Password:  nil,
	}

	conn, err := player.CreateRoomAndJoin(ctx, room)
	g.Expect(err).ShouldNot(HaveOccurred())

	msg, err := conn.Sock().Recv()
	g.Expect(err).ShouldNot(HaveOccurred())

	switch msg := msg.Message.(type) {
	case *gamesvc.Message_UpdateRoom:
		msg.UpdateRoom.Room.Id = ""

		g.Expect(msg.UpdateRoom).To(matcher.EqualCmp(&gamesvc.UpdateRoom{
			Room: &gamesvc.Room{
				Id:        "",
				Name:      room.Name,
				LeaderId:  player.Proto().Id,
				IsPublic:  room.IsPublic,
				Langugage: room.Langugage,
				Lobby: []*gamesvc.Player{
					{
						Id:          player.Proto().Id,
						Name:        player.Proto().Name,
						GravatarUrl: player.Proto().GravatarUrl,
					},
				},
				Teams: []*gamesvc.Team{},
			},
			Password: nil,
		}))
	default:
		t.Fatalf("unexpected first message type: %T", msg)
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
			Id:        "",
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
			update.UpdateRoom.Room.Id = ""

			return update.UpdateRoom
		}).WithContext(ctx).Should(matcher.EqualCmp(roomMsg))
	}
}
