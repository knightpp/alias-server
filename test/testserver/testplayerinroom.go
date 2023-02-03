package testserver

import (
	"fmt"
	"reflect"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
)

type TestPlayerInRoom struct {
	sock      gamesvc.GameService_JoinClient
	authToken string
	player    *gamesvc.Player
	cancel    func()
}

func (ctp *TestPlayerInRoom) ID() string {
	return ctp.player.Id
}

func (ctp *TestPlayerInRoom) Sock() gamesvc.GameService_JoinClient {
	return ctp.sock
}

func (ctp *TestPlayerInRoom) Cancel() {
	ctp.cancel()
}

func (tpr *TestPlayerInRoom) RecvAndAssert(out any) error {
	msg, err := tpr.sock.Recv()
	if err != nil {
		return fmt.Errorf("recv msg: %w", err)
	}

	expected := reflect.TypeOf(out)
	actual := reflect.TypeOf(msg.Message)

	if expected != actual {
		return fmt.Errorf("expected: %s, actual: %s", expected, actual)
	}

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(msg.Message).Elem())

	return nil
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
