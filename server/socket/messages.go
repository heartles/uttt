package socket

import (
	"encoding/json"

	"github.com/heartles/uttt/server/game"
	"github.com/heartles/uttt/server/store"
)

type IncomingSocketMessage struct {
	Type    string          `json:"messageType"`
	Payload json.RawMessage `json:"payload"`
}

type OutgoingSocketMessage struct {
	Type    string      `json:"messageType"`
	Payload interface{} `json:"payload"`
}

type LoginRequest struct {
	LoginID string `json:"loginID"`
}

type NewGame struct {
	Opponent string `json:"opponent"`
}

type PlayMove struct {
	GameID string    `json:"gameID"`
	Move   game.Move `json:"move"`
}

type LoginSuccess struct {
	Username string            `json:"username"`
	PlayerID string            `json:"playerID"`
	Games    []store.GameState `json:"games"`
}

type ErrorMessage struct {
	Message string `json:"message"`

	// if Recoverable is false, then the websocket is closed after this
	// message is sent
	Recoverable bool `json:"recoverable"`
}
