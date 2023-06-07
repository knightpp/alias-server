package socket_test

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	gameserver "github.com/knightpp/alias-server/internal/gameserver"
	"github.com/knightpp/alias-server/internal/storage/memory"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestTwo(t *testing.T) {
	g := NewWithT(t)
	log := zerolog.New(zerolog.NewTestWriter(t))
	ctx := context.Background()
	if time, ok := t.Deadline(); ok {
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, time)

		defer cancel()
	}

	player1Token := "player1"
	player1 := protoPlayer(1)
	player2Token := "player2"
	player2 := protoPlayer(2)
	mem := memory.New()

	err := mem.SetPlayer(ctx, player1Token, player1)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mem.SetPlayer(ctx, player2Token, player2)
	g.Expect(err).ShouldNot(HaveOccurred())

	var address string
	{
		listener, err := net.Listen("tcp", "127.0.0.1:")
		g.Expect(err).ShouldNot(HaveOccurred())

		address = listener.Addr().String()

		authFunc := func(ctx context.Context) (context.Context, error) {
			tokenParts := metadata.ValueFromIncomingContext(ctx, "token")
			token := strings.Join(tokenParts, "")
			player, err := mem.GetPlayer(ctx, token)
			if err != nil {
				return nil, err
			}

			return context.WithValue(ctx, gameserver.AuthKey{}, player), nil
		}
		service := gameserver.New(log, mem)
		server := grpc.NewServer(
			grpc.StreamInterceptor(auth.StreamServerInterceptor(authFunc)),
			grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authFunc)),
		)
		server.RegisterService(&gamesvc.GameService_ServiceDesc, service)

		go func() {
			err := server.Serve(listener)
			g.Expect(err).ShouldNot(HaveOccurred())
		}()

		t.Cleanup(func() {
			server.GracefulStop()
		})
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "token", player1Token)

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	g.Expect(err).ShouldNot(HaveOccurred())

	client := gamesvc.NewGameServiceClient(conn)

	resp, err := client.CreateRoom(ctx, protoRoom())
	g.Expect(err).ShouldNot(HaveOccurred())

	sub, err := client.JoinRoom(ctx, &gamesvc.JoinRoomRequest{
		RoomId: resp.Id,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	ann, err := sub.Recv()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(ann.Announcement.GetAddPlayer()).Should(Equal(&gamesvc.AnnAddPlayer{
		Player: nil,
	}))
}
