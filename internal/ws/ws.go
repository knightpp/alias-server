package ws

import (
	"fmt"

	"github.com/gorilla/websocket"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"google.golang.org/protobuf/proto"
)

type Conn struct {
	conn *websocket.Conn
}

func Wrap(conn *websocket.Conn) *Conn {
	return &Conn{
		conn: conn,
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) SendPlayerJoined(joined *serverpb.PlayerJoinedMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Joined{
			Joined: joined,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (c *Conn) SendPlayerLeft(left *serverpb.PlayerLeftMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Left{
			Left: left,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (c *Conn) SendWords(words *serverpb.WordsMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Words{
			Words: words,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (c *Conn) SendFatal(fatal *serverpb.FatalMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Fatal{
			Fatal: fatal,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (c *Conn) SendInitRoom(initRoom *serverpb.InitRoomMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_InitRoom{
			InitRoom: initRoom,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}
