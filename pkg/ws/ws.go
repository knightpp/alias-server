package ws

import (
	"fmt"

	"github.com/gorilla/websocket"
	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"google.golang.org/protobuf/proto"
)

//go:generate mockery --name Conn --with-expecter

type Conn interface {
	Close() error

	ReceiveMessage() (*serverpb.Message, error)

	SendPlayerJoined(joined *serverpb.PlayerJoinedMessage) error
	SendPlayerLeft(left *serverpb.PlayerLeftMessage) error
	SendWords(words *serverpb.WordsMessage) error
	SendFatal(fatal *serverpb.FatalMessage) error
	SendInitRoom(initRoom *serverpb.InitRoomMessage) error
	SendTeam(newTeam *serverpb.TeamMessage) error
	SendError(msg string) error
}

type wrapper struct {
	conn *websocket.Conn
}

func Wrap(conn *websocket.Conn) Conn {
	return &wrapper{
		conn: conn,
	}
}

func (c *wrapper) Close() error {
	return c.conn.Close()
}

func (c wrapper) ReceiveMessage() (*serverpb.Message, error) {
	mt, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read websocket: %w", err)
	}

	if mt != websocket.BinaryMessage {
		return nil, fmt.Errorf("expected binary message")
	}

	var msg serverpb.Message

	err = proto.Unmarshal(data, &msg)
	return &msg, err
}

func (c wrapper) SendPlayerJoined(joined *serverpb.PlayerJoinedMessage) error {
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

func (c wrapper) SendPlayerLeft(left *serverpb.PlayerLeftMessage) error {
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

func (c wrapper) SendWords(words *serverpb.WordsMessage) error {
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

func (c wrapper) SendFatal(fatal *serverpb.FatalMessage) error {
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

func (c wrapper) SendInitRoom(initRoom *serverpb.InitRoomMessage) error {
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

func (c wrapper) SendTeam(newTeam *serverpb.TeamMessage) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Team{
			Team: newTeam,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}

func (c wrapper) SendError(msg string) error {
	msgBytes, err := proto.Marshal(&serverpb.Message{
		Message: &serverpb.Message_Error{
			Error: &serverpb.ErrorMessage{
				Error: msg,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("marshal proto: %w", err)
	}

	err = c.conn.WriteMessage(websocket.BinaryMessage, msgBytes)
	return err
}
