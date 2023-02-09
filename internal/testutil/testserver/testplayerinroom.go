package testserver

import (
	"fmt"
	"sync"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
)

type TestPlayerInRoom struct {
	C chan *gamesvc.Message

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
			return fmt.Errorf("recv msg: %w", err)
		}

		select {
		case <-ctp.done:
			close(ctp.C)
			return nil

		case ctp.C <- msg:
			continue
		}
	}
}

func (ctp *TestPlayerInRoom) Next() *gamesvc.Message {
	msg, ok := <-ctp.C
	if !ok {
		panic("channel was closed")
	}

	return msg
}

func (ctp *TestPlayerInRoom) Poll() *gamesvc.Message {
	select {
	case msg, ok := <-ctp.C:
		if !ok {
			panic("channel was closed")
		}

		if msg, ok := msg.Message.(*gamesvc.Message_Error); ok {
			panic(msg.Error.Error)
		}

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

func (tpr *TestPlayerInRoom) StartGame() error {
	return tpr.sock.Send(&gamesvc.Message{
		Message: &gamesvc.Message_Start{
			Start: &gamesvc.MsgStart{},
		},
	})
}
