package actor

import (
	"reflect"
	"sync"
	"testing"

	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewRoomFromRequest(t *testing.T) {
	type args struct {
		req       *serverpb.CreateRoomRequest
		creatorID string
	}
	tests := []struct {
		name string
		args args
		want *Room
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRoomFromRequest(tt.args.req, tt.args.creatorID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRoomFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoom_RemovePlayer(t *testing.T) {
	type fields struct {
		Id       string
		Name     string
		LeaderId string
		IsPublic bool
		Language string
		Lobby    map[string]*Player
		Teams    map[string]*Team
		Password *string
	}
	type args struct {
		playerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "when there is no such user in lobby",
			fields: fields{
				Lobby: map[string]*Player{},
			},
			args: args{
				playerID: "abcd",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Room{
				Id:       tt.fields.Id,
				Name:     tt.fields.Name,
				LeaderId: tt.fields.LeaderId,
				IsPublic: tt.fields.IsPublic,
				Language: tt.fields.Language,
				Lobby:    tt.fields.Lobby,
				Teams:    tt.fields.Teams,
				Password: tt.fields.Password,
				mutex:    sync.Mutex{},
				log:      zerolog.Nop(),
			}

			err := r.RemovePlayer(tt.args.playerID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// func TestRoom_AddPlayerToLobby(t *testing.T) {
// 	type fields struct {
// 		Id       string
// 		Name     string
// 		LeaderId string
// 		IsPublic bool
// 		Language string
// 		Lobby    map[string]*Player
// 		Teams    map[string]*Team
// 		Password *string
// 		mutex    sync.Mutex
// 		log      zerolog.Logger
// 	}
// 	type args struct {
// 		p *Player
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &Room{
// 				Id:       tt.fields.Id,
// 				Name:     tt.fields.Name,
// 				LeaderId: tt.fields.LeaderId,
// 				IsPublic: tt.fields.IsPublic,
// 				Language: tt.fields.Language,
// 				Lobby:    tt.fields.Lobby,
// 				Teams:    tt.fields.Teams,
// 				Password: tt.fields.Password,
// 				mutex:    tt.fields.mutex,
// 				log:      tt.fields.log,
// 			}
// 			if err := r.AddPlayerToLobby(tt.args.p); (err != nil) != tt.wantErr {
// 				t.Errorf("Room.AddPlayerToLobby() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestRoom_CreateTeam(t *testing.T) {
// 	type fields struct {
// 		Id       string
// 		Name     string
// 		LeaderId string
// 		IsPublic bool
// 		Language string
// 		Lobby    map[string]*Player
// 		Teams    map[string]*Team
// 		Password *string
// 		mutex    sync.Mutex
// 		log      zerolog.Logger
// 	}
// 	type args struct {
// 		req *serverpb.CreateTeamRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *Team
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &Room{
// 				Id:       tt.fields.Id,
// 				Name:     tt.fields.Name,
// 				LeaderId: tt.fields.LeaderId,
// 				IsPublic: tt.fields.IsPublic,
// 				Language: tt.fields.Language,
// 				Lobby:    tt.fields.Lobby,
// 				Teams:    tt.fields.Teams,
// 				Password: tt.fields.Password,
// 				mutex:    tt.fields.mutex,
// 				log:      tt.fields.log,
// 			}
// 			got, err := r.CreateTeam(tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Room.CreateTeam() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Room.CreateTeam() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestRoom_forEachPlayer(t *testing.T) {
// 	type fields struct {
// 		Id       string
// 		Name     string
// 		LeaderId string
// 		IsPublic bool
// 		Language string
// 		Lobby    map[string]*Player
// 		Teams    map[string]*Team
// 		Password *string
// 		mutex    sync.Mutex
// 		log      zerolog.Logger
// 	}
// 	type args struct {
// 		f func(p *Player) error
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &Room{
// 				Id:       tt.fields.Id,
// 				Name:     tt.fields.Name,
// 				LeaderId: tt.fields.LeaderId,
// 				IsPublic: tt.fields.IsPublic,
// 				Language: tt.fields.Language,
// 				Lobby:    tt.fields.Lobby,
// 				Teams:    tt.fields.Teams,
// 				Password: tt.fields.Password,
// 				mutex:    tt.fields.mutex,
// 				log:      tt.fields.log,
// 			}
// 			if err := r.forEachPlayer(tt.args.f); (err != nil) != tt.wantErr {
// 				t.Errorf("Room.forEachPlayer() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
