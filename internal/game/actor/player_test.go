package actor

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	"github.com/knightpp/alias-server/internal/gravatar"
	connmock "github.com/knightpp/alias-server/internal/ws/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var playerID = uuid.New().String()

func TestCreateTeam(t *testing.T) {
	assert := assert.New(t)
	conn := connmock.NewConn(t)
	player := NewPlayerFromPB(&modelpb.Player{
		Id:          playerID,
		Name:        "my test",
		GravatarUrl: gravatar.GetUrlOrDefault(nil),
	}, conn)
	testErr := errors.New("test error")
	conn.EXPECT().ReceiveMessage().Return(nil, testErr)

	err := player.RunLoop(zerolog.Nop())

	assert.ErrorIs(err, testErr)
}
