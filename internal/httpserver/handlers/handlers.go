// Package handlers contains all http handlers for server
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ole-larsen/binance-subscriber/internal/storage"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

type StatusResponse struct {
	Status string `json:"status"`
}

// Status godoc
// @Tags Info
// @Summary server status
// @ID serverStatus
// @Accept  json
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /status [get].
func StatusHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	_, err := rw.Write([]byte(`{"status":"ok"}`))
	if err != nil {
		InternalServerErrorRequest(rw, r)
		return
	}
}

func BadRequest(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(rw, fmt.Sprintf("%d", http.StatusBadRequest)+" bad request", http.StatusBadRequest)
}

func NotFoundRequest(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(rw, fmt.Sprintf("%d", http.StatusNotFound)+" page not found", http.StatusNotFound)
}

func NotAllowedRequest(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(rw, fmt.Sprintf("%d", http.StatusMethodNotAllowed)+" method not allowed", http.StatusMethodNotAllowed)
}

func ForbiddenRequest(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(rw, fmt.Sprintf("%d", http.StatusForbidden)+" forbidden", http.StatusForbidden)
}

func InternalServerErrorRequest(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.Error(rw, fmt.Sprintf("%d", http.StatusInternalServerError)+" internal server error", http.StatusInternalServerError)
}

// WebSocketHandler godoc
// @Tags WebSocket
// @Summary Handle WebSocket connections
// @ID websocketConnection
// @Accept  json
// @Produce json
// @Router /ws [get]

func WebSocketHandler(store storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if store == nil {
			InternalServerErrorRequest(rw, r)
			return
		}

		data := store.GetAll()

		dataStr, err := json.Marshal(data)

		if err != nil {
			InternalServerErrorRequest(rw, r)
			return
		}

		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			InternalServerErrorRequest(rw, r)
			return
		}
		defer conn.Close()

		if err := conn.WriteMessage(websocket.TextMessage, dataStr); err != nil {
			InternalServerErrorRequest(rw, r)
			return
		}

		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					break
				}

				InternalServerErrorRequest(rw, r)

				return
			}

			data := store.GetAll()

			dataStr, err := json.Marshal(data)

			if err != nil {
				InternalServerErrorRequest(rw, r)
				return
			}

			if err := conn.WriteMessage(messageType, dataStr); err != nil {
				InternalServerErrorRequest(rw, r)
				return
			}
		}
	}
}
