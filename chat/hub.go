package chat

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // for testing purposes, allow all origins
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// WriteBufferPool: &poolAdapter{bufferPool}}
}

const ServerID string = "serverA"

// type ChatManager struct {
// 	sockets  map[string]*websocket.Conn
// 	mu       sync.Mutex
// 	handlers map[string]EventHandler
// }

// func NewChatManager() *ChatManager {
// 	m := &ChatManager{
// 		sockets:  make(map[string]*websocket.Conn),
// 		handlers: make(map[string]EventHandler),
// 	}
// 	m.setupEventHandlers()
// 	return m
// }

// func (m *ChatManager) setupEventHandlers() {
// 	m.handlers[EventSendMessage] = SendMessage
// }

// func SendMessage(event Event, c *Client) error {
// 	fmt.Println("*** Entered in SendMessage function call")
// 	fmt.Println("*** Printing the received message", event)
// 	n := map[string]string{
// 		"name": "ajay",
// 		"city": "dehradun",
// 	}
// 	payloadBytes, _ := json.Marshal(n)
// 	ne := Event{
// 		Type:    "send_message",
// 		Payload: json.RawMessage(payloadBytes),
// 	}
// 	fmt.Println("*** Sending the data back to the channel")
// 	c.msgCH <- ne
// 	return nil
// }

// func (m *ChatManager) routeEvent(event Event, c *Client) error {
// 	if handler, ok := m.handlers[event.Type]; ok {
// 		if err := handler(event, c); err != nil {
// 			return err
// 		}
// 		return nil
// 	} else {
// 		fmt.Println("There is no such event")
// 		return fmt.Errorf("There is no such event")
// 	}
// }

// func (cm *ChatManager) RegisterClient(client *Client) {
// 	cm.mu.Lock()
// 	cm.sockets[client.userID] = client.conn
// 	cm.mu.Unlock()
// }

// func (cm *ChatManager) UnregisterClient(client *Client) {
// 	cm.mu.Lock()
// 	defer cm.mu.Unlock()

// 	if _, ok := cm.sockets[client.userID]; ok {
// 		delete(cm.sockets, client.userID)
// 	}
// }

func NewHub() *Hub {
	hub := &Hub{
		ID:      ServerID,
		Clients: make(map[string]*Client),
	}
	return hub
}
