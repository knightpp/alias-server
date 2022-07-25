package game

import (
	"reflect"
	"testing"

	modelpb "github.com/knightpp/alias-proto/go/pkg/model/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/actor"
	"github.com/knightpp/alias-server/internal/storage"
	dbmock "github.com/knightpp/alias-server/internal/storage/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	log := zerolog.Nop()
	db := dbmock.NewPlayerDB(t)
	rooms := fp.NewLocker(make(map[string]*actor.Room))
	assert := assert.New(t)

	game := New(log, db)

	assert.NotNil(game)
	assert.Equal(log, game.log)
	assert.Equal(db, game.playerDB)
	assert.Equal(rooms, game.rooms)
}

func TestGame_CreateRoom(t *testing.T) {
	type fields struct {
		rooms map[string]*actor.Room
	}
	type args struct {
		room *actor.Room
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *actor.Room
		wantErr bool
	}{
		{
			name: "when room already exists",
			fields: fields{
				rooms: map[string]*actor.Room{
					"test-id-1": {
						Id: "test-id-1",
					},
				},
			},
			args: args{
				room: &actor.Room{
					Id: "test-id-1",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "when room creation succeeds",
			fields: fields{
				rooms: map[string]*actor.Room{
					"test-id-1": {
						Id: "test-id-1",
					},
				},
			},
			args: args{
				room: &actor.Room{
					Id: "test-id-2",
				},
			},
			want: func() *actor.Room {
				r := &actor.Room{
					Id: "test-id-2",
				}
				r.SetLogger(zerolog.Nop())
				return r
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(zerolog.Nop(), dbmock.NewPlayerDB(t))
			if tt.fields.rooms != nil {
				g.rooms = fp.NewLocker(tt.fields.rooms)
			}

			got, err := g.CreateRoom(tt.args.room)

			if (err != nil) != tt.wantErr {
				t.Errorf("Game.CreateRoom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				got.SetLogger(zerolog.Nop())
			}
			assert.Equal(t, tt.want, got, "Game.CreateRoom()")
		})
	}
}

func TestGame_ListRooms(t *testing.T) {
	type fields struct {
		rooms map[string]*actor.Room
	}
	tests := []struct {
		name   string
		fields fields
		want   []*actor.Room
	}{
		{
			name: "when there are no rooms",
			fields: fields{
				rooms: map[string]*actor.Room{},
			},
			want: nil,
		},
		{
			name: "when there are two rooms",
			fields: fields{
				rooms: map[string]*actor.Room{
					"id-1": {Id: "id-1"},
					"id-2": {Id: "id-2"},
				},
			},
			want: []*actor.Room{{Id: "id-1"}, {Id: "id-2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(zerolog.Nop(), dbmock.NewPlayerDB(t))
			if tt.fields.rooms != nil {
				g.rooms = fp.NewLocker(tt.fields.rooms)
			}

			got := g.ListRooms()

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestGame_GetRoom(t *testing.T) {
	type fields struct {
		log      zerolog.Logger
		rooms    fp.Locker[map[string]*actor.Room]
		playerDB storage.PlayerDB
	}
	type args struct {
		roomID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *actor.Room
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				log:      tt.fields.log,
				rooms:    tt.fields.rooms,
				playerDB: tt.fields.playerDB,
			}
			got, got1 := g.GetRoom(tt.args.roomID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Game.GetRoom() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Game.GetRoom() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGame_GetPlayerInfo(t *testing.T) {
	type fields struct {
		log      zerolog.Logger
		rooms    fp.Locker[map[string]*actor.Room]
		playerDB storage.PlayerDB
	}
	type args struct {
		playerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *modelpb.Player
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				log:      tt.fields.log,
				rooms:    tt.fields.rooms,
				playerDB: tt.fields.playerDB,
			}
			got, err := g.GetPlayerInfo(tt.args.playerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Game.GetPlayerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Game.GetPlayerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
