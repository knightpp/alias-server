package game

import (
	"errors"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	"github.com/knightpp/alias-server/internal/fp"
	"github.com/knightpp/alias-server/internal/game/entity"
	"github.com/knightpp/alias-server/internal/uuidgen"
	"github.com/lrita/cmap"
	"github.com/rs/zerolog"
)

var (
	ErrRoomNoTeams        = entity.ErrStartNoTeams
	ErrRoomIncompleteTeam = entity.ErrStartIncompleteTeam
	ErrAlreadyInRoom      = errors.New("you must first leave the previous room to join another")
	ErrNotJoinedRoom      = errors.New("you must first join a room")
	ErrNotInRoom          = errors.New("you are not in room")
	ErrRoomNotFound       = errors.New("room not found")
	ErrPlayerInRoom       = errors.New("player already in the room")
	ErrAlreadySubscribed  = errors.New("you are already subscribed")
)

type Game struct {
	log zerolog.Logger

	playerToRoom cmap.Map[string, string]
	rooms        cmap.Map[string, *entity.Room]
}

func New(log zerolog.Logger) *Game {
	return &Game{
		log: log,
	}
}

func (g *Game) CreateRoom(
	leader *gamesvc.Player,
	req *gamesvc.CreateRoomRequest,
) (roomID string) {
	roomID = uuidgen.NewString()
	r := entity.NewRoom(g.log, roomID, leader.Id, req)

	g.rooms.Store(roomID, r)

	return roomID
}

func (g *Game) ListRooms() []*gamesvc.Room {
	roomsProto := make([]*gamesvc.Room, 0, g.rooms.Count())
	g.rooms.Range(func(_ string, room *entity.Room) bool {
		room.Mutex.Lock()
		defer room.Mutex.Unlock()

		proto := room.GetProto()
		// returns nil if room was deleted from map
		if proto == nil {
			return true
		}

		roomsProto = append(roomsProto, proto)
		return true
	})
	return roomsProto
}

func (g *Game) JoinRoom(
	player *gamesvc.Player,
	req *gamesvc.JoinRoomRequest,
	srv gamesvc.GameService_JoinRoomServer,
) error {
	room, ok := g.rooms.Load(req.RoomId)
	if !ok {
		return ErrRoomNotFound
	}

	_, loaded := g.playerToRoom.LoadOrStore(player.Id, req.RoomId)
	if loaded {
		return ErrAlreadyInRoom
	}

	room.Mutex.Lock()

	if room.HasPlayer(player.Id) {
		room.Mutex.Unlock()
		return ErrPlayerInRoom
	}

	playerEnt := entity.NewPlayer(g.log, srv, player)

	room.Lobby = append(room.Lobby, playerEnt)

	err := room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_AddPlayer{
			AddPlayer: &gamesvc.AnnAddPlayer{
				Player: player,
			},
		},
	})
	if err != nil {
		room.Mutex.Unlock()
		return err
	}
	room.Mutex.Unlock()

	<-playerEnt.Done()

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	_, ok = room.RemovePlayer(player.Id)
	if ok {
		if room.IsEmpty() {
			g.rooms.Delete(room.Id)
			return nil
		}

		return room.Announce(&gamesvc.Announcement{
			Announce: &gamesvc.Announcement_RemovePlayer{
				RemovePlayer: &gamesvc.AnnRemovePlayer{
					PlayerId: player.Id,
				},
			},
		})
	}

	return nil
}

func (g *Game) TransferLeadership(player *gamesvc.Player, req *gamesvc.TransferLeadershipRequest) (*gamesvc.TransferLeadershipResponse, error) {
	if player.Id == req.PlayerId {
		return nil, errors.New("could not transfer leadership to yourself")
	}

	room, err := g.loadRoom(req.PlayerId)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	exists := room.HasPlayer(req.PlayerId)
	if !exists {
		return nil, fmt.Errorf("could not transfer leadership: no player with id=%s", req.PlayerId)
	}

	room.LeaderId = req.PlayerId
	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_TransferLeadership{
			TransferLeadership: &gamesvc.AnnTransferLeadership{
				NewLeader: req.PlayerId,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &gamesvc.TransferLeadershipResponse{}, nil
}

func (g *Game) CreateTeam(player *gamesvc.Player, req *gamesvc.CreateTeamRequest) (*gamesvc.CreateTeamResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	team := &entity.Team{
		ID:      uuid.NewString(),
		Name:    req.Name,
		PlayerA: nil,
		PlayerB: nil,
	}
	if team.Name == "" {
		team.Name = gofakeit.Vegetable()
	}

	room.Teams = append(room.Teams, team)

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_NewTeam{
			NewTeam: &gamesvc.AnnNewTeam{
				Team: team.ToProto(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &gamesvc.CreateTeamResponse{}, nil
}

// func (g *Game) UpdateTeam(player *gamesvc.Player, req *gamesvc.UpdateTeamRequest) (*gamesvc.UpdateTeamResponse, error)

func (g *Game) JoinTeam(player *gamesvc.Player, req *gamesvc.JoinTeamRequest) (*gamesvc.JoinTeamResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	team, ok := fp.Find(room.Teams, func(t *entity.Team) bool {
		return t.ID == req.TeamId
	})
	if !ok {
		return nil, fmt.Errorf("TODO: team not found")
	}

	removedPlayer, ok := room.RemovePlayer(player.Id)
	if !ok {
		return nil, errors.New("player not found")
	}

	if team.PlayerA != nil && team.PlayerB != nil {
		return nil, errors.New("team is full")
	}

	var slot gamesvc.Slot
	switch {
	case team.PlayerA == nil:
		team.PlayerA = removedPlayer
		slot = gamesvc.Slot_SLOT_A
	case team.PlayerB == nil:
		team.PlayerB = removedPlayer
		slot = gamesvc.Slot_SLOT_B
	default:
		return nil, fmt.Errorf("TODO: team is full")
	}

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_JoinTeam{
			JoinTeam: &gamesvc.AnnJoinTeam{
				PlayerId: player.Id,
				TeamId:   team.ID,
				Slot:     slot,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &gamesvc.JoinTeamResponse{}, nil
}

func (g *Game) StartGame(player *gamesvc.Player, req *gamesvc.StartGameRequest) (*gamesvc.StartGameResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if room.LeaderId != player.Id {
		return nil, errors.New("only leader id can start game")
	}

	if len(room.Teams) == 0 {
		return nil, entity.ErrStartNoTeams
	}

	playerIDTurn := req.NextPlayerTurn
	if playerIDTurn == "" {
		return nil, fmt.Errorf("next player turn should not be empty")
	}

	for _, team := range room.Teams {
		if team.PlayerA == nil || team.PlayerB == nil {
			return nil, entity.ErrStartIncompleteTeam
		}
	}

	ok := room.HasPlayer(playerIDTurn)
	if !ok {
		return nil, fmt.Errorf("cannot start game: no player with %q id", playerIDTurn)
	}

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_StartGame{
			StartGame: &gamesvc.AnnStartGame{
				PlayerTurn: playerIDTurn,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &gamesvc.StartGameResponse{}, nil
}

func (g *Game) StopGame(player *gamesvc.Player, req *gamesvc.StopGameRequest) (*gamesvc.StopGameResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if room.LeaderId != player.Id {
		return nil, errors.New("only leader can end game")
	}

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_EndGame{
			EndGame: &gamesvc.AnnEndGame{
				TeamIdToStats: room.TotalStats,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &gamesvc.StopGameResponse{}, nil
}

func (g *Game) StartTurn(player *gamesvc.Player, req *gamesvc.StartTurnRequest) (*gamesvc.StartTurnResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if player.Id != room.PlayerIDTurn {
		return nil, fmt.Errorf("only player with %s id can start next turn", room.PlayerIDTurn)
	}
	if req.DurationMs == 0 {
		return nil, errors.New("could not start turn with 0 duration")
	}

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_StartTurn{
			StartTurn: &gamesvc.AnnStartTurn{
				DurationMs: req.DurationMs,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	room.TurnDeadline = time.Now().Add(time.Duration(req.DurationMs) * time.Millisecond)
	return &gamesvc.StartTurnResponse{}, nil
}

func (g *Game) StopTurn(player *gamesvc.Player, req *gamesvc.StopTurnRequest) (*gamesvc.StopTurnResponse, error) {
	room, err := g.loadRoom(player.Id)
	if err != nil {
		return nil, err
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if player.Id != room.PlayerIDTurn {
		return nil, fmt.Errorf("only %q player can end turn", room.PlayerIDTurn)
	}

	err = room.Announce(&gamesvc.Announcement{
		Announce: &gamesvc.Announcement_EndTurn{
			EndTurn: &gamesvc.AnnEndTurn{
				Stats: room.TurnStats,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	team, ok := room.FindTeamWithPlayer(player.Id)
	if ok {
		prevStats, ok := room.TotalStats[team.ID]
		if ok {
			prevStats.Rights += room.TurnStats.Rights
			prevStats.Wrongs += room.TurnStats.Wrongs
		} else {
			prevStats = room.TurnStats
		}

		room.TotalStats[team.ID] = prevStats
	}

	return &gamesvc.StopTurnResponse{}, nil
}

// func (g *Game) Score(player *gamesvc.Player, req *gamesvc.ScoreRequest) (*gamesvc.ScoreResponse, error)

// loadRoom returns room in which player are in.
func (g *Game) loadRoom(playerID string) (*entity.Room, error) {
	roomID, ok := g.playerToRoom.Load(playerID)
	if !ok {
		return nil, ErrNotJoinedRoom
	}

	room, ok := g.rooms.Load(roomID)
	if !ok {
		return nil, ErrRoomNotFound
	}

	return room, nil
}
