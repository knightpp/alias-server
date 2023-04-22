package testserver

import (
	"context"
	"sync"
	"time"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	"github.com/rs/zerolog"
)

type TestPlayerInRoom struct {
	C      chan *gamesvc.Message
	logger zerolog.Logger

	sock   gamesvc.GameService_JoinClient
	player *gamesvc.Player

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
		return msg
	}
}

func (ctp *TestPlayerInRoom) NextMsgUnpack(ctx context.Context) any {
	select {
	case <-ctx.Done():
		return nil
	case msg := <-ctp.C:
		switch msg := msg.Message.(type) {
		case *gamesvc.Message_Error:
			return msg.Error
		case *gamesvc.Message_UpdateRoom:
			return msg.UpdateRoom
		case *gamesvc.Message_TransferLeadership:
			return msg.TransferLeadership
		case *gamesvc.Message_CreateTeam:
			return msg.CreateTeam
		case *gamesvc.Message_TeamCreated:
			return msg.TeamCreated
		case *gamesvc.Message_JoinTeam:
			return msg.JoinTeam
		case *gamesvc.Message_StartGame:
			return msg.StartGame
		case *gamesvc.Message_EndGame:
			return msg.EndGame
		case *gamesvc.Message_StartTurn:
			return msg.StartTurn
		case *gamesvc.Message_EndTurn:
			return msg.EndTurn
		case *gamesvc.Message_Results:
			return msg.Results
		case *gamesvc.Message_Word:
			return msg.Word
		default:
			return msg
		}
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

func (ctp *TestPlayerInRoom) CreateTeam(name string) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_CreateTeam{
			CreateTeam: &gamesvc.MsgCreateTeam{
				Name: name,
			},
		},
	})
}

func (ctp *TestPlayerInRoom) JoinTeam(id string) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_JoinTeam{
			JoinTeam: &gamesvc.MsgJoinTeam{
				TeamId: id,
			},
		},
	})
}

func (ctp *TestPlayerInRoom) TransferLeadership(playerID string) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_TransferLeadership{
			TransferLeadership: &gamesvc.MsgTransferLeadership{
				PlayerId: playerID,
			},
		},
	})
}

func (ctp *TestPlayerInRoom) StartGame(nextPlayerIDTurn string) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_StartGame{
			StartGame: &gamesvc.MsgStartGame{
				NextPlayerTurn: nextPlayerIDTurn,
			},
		},
	})
}

func (ctp *TestPlayerInRoom) StartTurn(duration time.Duration) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_StartTurn{
			StartTurn: &gamesvc.MsgStartTurn{
				DurationMs: uint64(duration.Milliseconds()),
			},
		},
	})
}

func (ctp *TestPlayerInRoom) EndTurn(rights, wrongs uint32) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_EndTurn{
			EndTurn: &gamesvc.MsgEndTurn{
				Stats: &gamesvc.Statistics{
					Rights: rights,
					Wrongs: wrongs,
				},
			},
		},
	})
}

func (ctp *TestPlayerInRoom) Word(word string) error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_Word{
			Word: &gamesvc.MsgWord{
				Word: word,
			},
		},
	})
}

func (ctp *TestPlayerInRoom) EndGame() error {
	return ctp.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_EndGame{EndGame: &gamesvc.MsgEndGame{}},
	})
}
