package chat

import (
	"fmt"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (cm *ChatManager) HandleChat(l *zap.Logger, ws *websocket.Conn, userID string) {
	wsClient := NewClient(userID, ws, cm)
	cm.RegisterClient(wsClient)
	l.Info(fmt.Sprintf("User %s connected >>>", userID))

	// goroutine process to read messages

	l.Info("Running read goroutine")
	go wsClient.readMessages(l)

	// goroutine process to write messages
	l.Info("Running write go routine")
	// go wsClient.writeMessage(l)
	wsClient.writeMessage(l)

}
