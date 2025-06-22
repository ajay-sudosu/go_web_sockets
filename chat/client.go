package chat

import (
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type Client struct {
	userID  string
	conn    *websocket.Conn
	manager *ChatManager
	msgCH   chan Event
}

func NewClient(userID string, conn *websocket.Conn, manager *ChatManager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		userID:  userID,
		msgCH:   make(chan Event),
	}
}

type ChatMessage struct {
	ToUser  string `json:"to_user"`
	Message string `json:"message"`
}

func (c *Client) readMessages(l *zap.Logger) {
	defer func() {
		l.Info("Defering the read message function and un registering the client")
		c.manager.UnregisterClient(c)
	}()

	for {
		l.Info("Reader is listining ....")
		_, r, err := c.conn.NextReader()
		if err != nil {
			l.Info("Error while connecting the reader")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.Info(fmt.Sprintf("This is the socket error: %s", err))
			} else {
				l.Info(fmt.Sprintf("WebSocket closed: %v\n", err))
			}
			break
		}
		message, err := io.ReadAll(r)
		if err != nil {
			l.Info(fmt.Sprintf("Error reading message: %v\n", err))
			break
		}
		var chatMsg Event
		err = json.Unmarshal(message, &chatMsg)
		if err != nil {
			fmt.Printf("Invalid message format: %v\n", err)
			continue
		}
		l.Info("Routing the event to its handler")
		if err := c.manager.routeEvent(chatMsg, c); err != nil {
			l.Info("Error in routing the message")
			return
		}
		l.Info("Routing the event is success")
	}
}

func (c *Client) writeMessage(l *zap.Logger) {
	defer func() {
		l.Info("Defering the write message function and un registering the client")
		c.manager.UnregisterClient(c)
	}()

	for {
		l.Info("Writer is listining ....")
		select {
		case msg, ok := <-c.msgCH:
			l.Info("Fetching the event...")
			if !ok {
				l.Info("Error in fetching the event from channel")
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			l.Info("Event fetched successfully marshaling the event")
			data, err := json.Marshal(msg)
			l.Info(fmt.Sprintf("*** this is the data to be written back to the channel: %s", data))
			if err != nil {
				l.Info(fmt.Sprintf("error marshaling message: %s", err))
				return
			}

			// targetConn, _ := c.manager.sockets[msg.ToUser]
			l.Info("Preparing write to channel")
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				l.Info(fmt.Sprintf("Error in getting the writer: %s", err))
			}
			l.Info("Writing the data back to channel")
			if _, err := writer.Write(data); err != nil {
				l.Info(fmt.Sprintf("error writing message:%s", err))
				// writer.Close() // ensure writer is closed even on error
			}
			l.Info("Message sent back")
			writer.Close()
		}
	}
}
