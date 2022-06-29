package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	playersWebsocketCurrent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "game_websocket_players",
		Help: "The current number of player in rooms",
	})

	playersWebsocketTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "game_websocket_players_total",
		Help: "The total number of players that joined a room",
	})

	roomsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "game_rooms_created_total",
		Help: "The total number of rooms created",
	})
)
