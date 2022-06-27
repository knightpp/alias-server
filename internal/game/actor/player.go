package actor

import (
	"fmt"

	"github.com/gorilla/websocket"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"google.golang.org/protobuf/proto"
)

type Player struct {
	Id          string
	Name        string
	GravatarUrl string

	conn *websocket.Conn
}

func NewPlayerFromPB(p *modelpb.Player, conn *websocket.Conn) Player {
	return Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
		conn:        conn,
	}
}

func (p Player) ToProto() *modelpb.Player {
	return &modelpb.Player{
		Id:          p.Id,
		Name:        p.Name,
		GravatarUrl: p.GravatarUrl,
	}
}

func (p Player) RunLoop() error {
	select {}
}

func (p Player) NotifyJoined(otherPlayer *modelpb.Player) error {
	msg := &serverpb.PlayerJoinedMessage{}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = p.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (p Player) NotifyLeft(playerID string) error {
	msg := &serverpb.PlayerLeftMessage{}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = p.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (p Player) send() {}
