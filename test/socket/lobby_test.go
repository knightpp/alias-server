package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/test/matcher"
	connector "github.com/knightpp/alias-server/test/testserver"
	. "github.com/onsi/gomega"
)

func TestJoinOneReceivesUpdateRoomWithTheirself(t *testing.T) {
	g := NewGomegaWithT(t)

	srv, err := connector.CreateAndStart(t)
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
