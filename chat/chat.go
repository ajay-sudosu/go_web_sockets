package chat

import (
	"fmt"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (h *Hub) HandleChat(l *zap.Logger, ws *websocket.Conn, rdbClient *RedisClientWrapper, userID string) {
	wsClient := NewClient(userID, ws, h, rdbClient, l)
	wsClient.RegisterClient(h, rdbClient)
	l.Info(fmt.Sprintf("User %s connected >>>", userID))

	// goroutine process to read messages
	l.Info("Running read goroutine")
	go wsClient.readMessages()

	// goroutine process to write messages
	l.Info("Running write go routine")
	// go wsClient.writeMessage(l)
	go wsClient.writeMessage()

}
