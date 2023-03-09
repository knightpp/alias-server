package testserver

import (
	"context"
	"sync"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

type TestPlayerInRoom struct {
	C         chan *gamesvc.Message
	RoomState *gamesvc.Room
	logger    zerolog.Logger

	sock      gamesvc.GameService_JoinClient
	authToken string
	player    *gamesvc.Player

	once   sync.Once
	done   chan struct{}
	cancel func()
}

func (ctp *TestPlayerInRoom) Start() error {
	for {
		msg, err := ctp.sock.Recv()
		if err != nil {
			return err
		}

		select {
		case <-ctp.done:
			close(ctp.C)
			return nil

		case ctp.C <- msg:
			ctp.logger.Info().Type("msg.type", msg.Message).Msg("received a message")
			continue
		}
	}
}

func (ctp *TestPlayerInRoom) NextMsg(ctx context.Context) *gamesvc.Message {
	select {
	case <-ctx.Done():
		return nil
	case msg := <-ctp.C:
		ctp.setRoomState(msg)
		return msg
	}
}

func (ctp *TestPlayerInRoom) setRoomState(msg *gamesvc.Message) {
	state, ok := msg.Message.(*gamesvc.Message_UpdateRoom)
	if !ok {
		return
	}

	ctp.RoomState = state.UpdateRoom.Room
}

func (ctp *TestPlayerInRoom) Poll(g Gomega) *gamesvc.Message {
	select {
	case msg, ok := <-ctp.C:
		if !ok {
			panic("channel was closed")
		}

		if msg, ok := msg.Message.(*gamesvc.Message_Error); ok {
			panic(msg.Error.Error)
			// g.ExpectWithOffset(1, msg.Error.Error).Should(BeEmpty())
		}

		ctp.setRoomState(msg)
		return msg
	default:
		return nil
	}
}

func (ctp *TestPlayerInRoom) PollRaw() *gamesvc.Message {
	select {
	case msg, ok := <-ctp.C:
		if !ok {
			panic("channel was closed")
		}

		ctp.setRoomState(msg)
		return msg
	default:
		return nil
	}
}

func (ctp *TestPlayerInRoom) ID() string {
	return ctp.player.Id
}

func (ctp *TestPlayerInRoom) Proto() *gamesvc.Player {
	return ctp.player
}

func (ctp *TestPlayerInRoom) Sock() gamesvc.GameService_JoinClient {
	return ctp.sock
}

func (ctp *TestPlayerInRoom) Cancel() {
	ctp.cancel()
	ctp.once.Do(func() {
		close(ctp.done)
	})
}

func (tpr *TestPlayerInRoom) CreateTeam(name string) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_CreateTeam{
			CreateTeam: &gamesvc.MsgCreateTeam{
				Name: name,
			},
		},
	})
}

func (tpr *TestPlayerInRoom) JoinTeam(id string) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_JoinTeam{
			JoinTeam: &gamesvc.MsgJoinTeam{
				TeamId: id,
			},
		},
	})
}

func (tpr *TestPlayerInRoom) TransferLeadership(playerID string) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_TransferLeadership{
			TransferLeadership: &gamesvc.MsgTransferLeadership{
				PlayerId: playerID,
			},
		},
	})
}

func (tpr *TestPlayerInRoom) StartGame(turns []string) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_StartGame{
			StartGame: &gamesvc.MsgStartGame{
				Turns: turns,
			},
		},
	})
}

func (tpr *TestPlayerInRoom) StartTurn(duration time.Duration) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_StartTurn{
			StartTurn: &gamesvc.MsgStartTurn{
				DurationMs: uint64(duration.Milliseconds()),
			},
		},
	})
}

func (tpr *TestPlayerInRoom) EndTurn(rights, wrongs uint32) error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_EndTurn{
			EndTurn: &gamesvc.MsgEndTurn{
				Rights: rights,
				Wrongs: wrongs,
			},
		},
	})
}
