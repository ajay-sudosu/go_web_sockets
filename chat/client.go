package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	SendCh chan []byte
	Hub    *Hub
	RDB    *RedisClientWrapper
	Logger *zap.Logger
}

type Hub struct {
	ID      string
	Clients map[string]*Client
	Mutex   sync.RWMutex
}

// func NewClient(userID string, conn *websocket.Conn, manager *ChatManager) *Client {
// 	return &Client{
// 		ID:     userID,
// 		Conn:   conn,
// 		SendCh: make(chan []byte),
// 	}
// }

func NewClient(userID string, conn *websocket.Conn, hub *Hub, rdb *RedisClientWrapper, logger *zap.Logger) *Client {
	return &Client{
		ID:     userID,
		Conn:   conn,
		SendCh: make(chan []byte, 256),
		Hub:    hub,
		RDB:    rdb,
		Logger: logger,
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.Logger.Info("Closing reader: unregistering client and closing connection")
		UnregisterClient(c.Hub, c, c.RDB)
		c.Conn.Close()
	}()
	c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		a := 10
		c.Logger.Debug(zap.String(a))
		c.Logger.Debug("Pong received, extending read deadline")
		return c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	})

	for {
		c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		c.Logger.Info("Reader listening for messages...")
		_, r, err := c.Conn.NextReader()
		if err != nil {
			c.Logger.Warn("WebSocket  reader error", zap.Error(err))
			break
		}

		message, err := io.ReadAll(r)
		if err != nil {
			c.Logger.Error("Failed to read message body", zap.Error(err))
			break
		}

		var chatMsg Event
		if err := json.Unmarshal(message, &chatMsg); err != nil {
			c.Logger.Warn("Invalid message format", zap.ByteString("raw", message))
			continue
		}

		c.Logger.Info("Received chat event", zap.Any("event", chatMsg))

		if err := c.Hub.RouteEvent(chatMsg, c); err != nil {
			c.Logger.Error("Failed to route event", zap.Error(err))
			break
		}
	}
}

func (c *Client) writeMessage() {
	defer func() {
		c.Logger.Info("Closing writer: unregistering client and closing connection")
		UnregisterClient(c.Hub, c, c.RDB)
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.SendCh:
			if !ok {
				c.Logger.Warn("Send channel closed, sending WebSocket close message")
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			writer, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.Logger.Error("Failed to get WebSocket writer", zap.Error(err))
				return
			}

			if _, err := writer.Write(msg); err != nil {
				c.Logger.Error("Failed to write message to WebSocket", zap.Error(err))
				writer.Close()
				return
			}

			writer.Close()
			c.Logger.Info("Message written to WebSocket", zap.ByteString("message", msg))
		}
	}
}

func (c *Client) RegisterClient(hub *Hub, rdb *RedisClientWrapper) {
	hub.Mutex.Lock()
	hub.Clients[c.ID] = c
	hub.Mutex.Unlock()

	err := rdb.Client.Set(context.Background(), c.ID, hub.ID, 0).Err()
	if err != nil {
		log.Println("Failed to register client in Redis:", err)
	}
}

func UnregisterClient(hub *Hub, client *Client, rdb *RedisClientWrapper) {
	hub.Mutex.Lock()
	delete(hub.Clients, client.ID)
	hub.Mutex.Unlock()

	err := rdb.Client.Del(context.Background(), "user:"+client.ID).Err()
	if err != nil {
		log.Println("Failed to unregister client in Redis:", err)
	}
}

func (h *Hub) RouteEvent(evt Event, from *Client) error {
	if evt.Type == EventSendMessage {
		toUser := evt.To
		message := evt.Payload

		// Check if recipient is connected locally (if present it will be found in hub)
		h.Mutex.RLock()
		target, ok := h.Clients[toUser]
		h.Mutex.RUnlock()

		if ok {
			// Deliver locally
			target.SendCh <- []byte(message)
			return nil
		}

		// Recipient is connected to a different server → publish to Redis
		payload := RPayload{
			To:      toUser,
			Message: message,
		}
		data, _ := json.Marshal(payload)
		toServerId, err := from.RDB.Client.Get(ctx, evt.To).Result()
		if err != nil {
			log.Println("Error getting server ID from Redis:", err)
			return err
		}
		return from.RDB.Client.Publish(ctx, toServerId, data).Err()
	}

	return fmt.Errorf("unknown event type: %s", evt.Type)
}
