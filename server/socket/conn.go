package socket

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/gorilla/websocket"
	"github.com/heartles/uttt/server/store"
)

var errMalformedRequest = errors.New("malformed request")
var errInvalidRequestType = errors.New("invalid request type")

type clientConn struct {
	socket    *websocket.Conn
	playerID  string
	openGames []store.NewGameNotification
	cancelCtx func()
	// todo: fix request id responses on server end
	currRequestID int
}

func (conn *clientConn) malformedRequest() error {
	return conn.sendError("malformed request", false)
}

func (conn *clientConn) sendError(message string, recoverable bool) error {
	err := conn.sendMessage(ErrorMessage{
		Message:     message,
		Recoverable: recoverable,
	})
	if !recoverable {
		conn.socket.Close()
	}
	return err
}

func (conn *clientConn) nextMessageSync() (interface{}, error) {
	_, reader, err := conn.socket.NextReader()
	if err != nil {
		return nil, err
	}

	return conn.parseMessage(reader)
}

func (conn *clientConn) sendMessage(payload interface{}) error {
	return conn.socket.WriteJSON(OutgoingSocketMessage{
		Type:      reflect.TypeOf(payload).Name(),
		Payload:   payload,
		RequestID: conn.currRequestID,
	})
}

func (conn *clientConn) parseMessage(r io.Reader) (interface{}, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var message IncomingSocketMessage
	err := decoder.Decode(&message)
	if err != nil {
		// TODO: log this
		return nil, errMalformedRequest
	}
	conn.currRequestID = message.RequestID

	decodedMessage := valueFromType(message.Type)
	if decodedMessage == nil {
		return nil, errInvalidRequestType
	}

	err = json.Unmarshal(message.Payload, decodedMessage)
	if err != nil {
		return nil, errMalformedRequest
	}

	return decodedMessage, nil
}

func valueFromType(typ string) interface{} {
	switch typ {
	case "LoginRequest":
		return &LoginRequest{}
	case "NewGame":
		return &NewGame{}
	case "UserLookup":
		return &UserLookup{}
	case "PlayMove":
		return &PlayMove{}
	}

	return nil
}
