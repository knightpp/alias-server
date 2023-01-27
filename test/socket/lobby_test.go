package socket_test

import (
	"context"
	"testing"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/knightpp/alias-server/internal/server"
	"github.com/knightpp/alias-server/internal/storage/mocks"
	"github.com/knightpp/alias-server/test/matcher"
	connector "github.com/knightpp/alias-server/test/testserver"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func TestJoinTwo(t *testing.T) {
	log := zerolog.New(zerolog.NewTestWriter(t))
	stormock := mocks.NewPlayer(t)

	gsvc := server.New(log, stormock)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"token": {"secret"},
	})

	room, err := gsvc.CreateRoom(ctx, &gamesvc.CreateRoomRequest{
		Name:      "test-room",
		IsPublic:  false,
		Langugage: "UA",
		Password:  nil,
	})
	if err != nil {
		panic(err)
	}

	sockmock := gamesvc.NewMockGameService_JoinServer(t)

	md, _ := metadata.FromIncomingContext(ctx)
	md.Append("room-id", room.Id)
	ctx = metadata.NewIncomingContext(ctx, md)

	sockmock.EXPECT().Context().Return(ctx).Once()
	stormock.EXPECT().GetPlayer(mock.Anything, "secret").Return(&gamesvc.Player{
		Id:   "id-1",
		Name: "Player name #1",
	}, nil).Once()
	sockmock.EXPECT().Recv().Return(&gamesvc.Message{
		Message: &gamesvc.Message_Error{
			Error: &gamesvc.MsgError{
				Error: "exit test",
			},
		},
	}, nil).Once()
	err = gsvc.Join(sockmock)
	if err != nil {
		panic(err)
	}
}

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
