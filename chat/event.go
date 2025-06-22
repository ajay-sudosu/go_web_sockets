package chat

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to send messages on the socket and depending on the type a particular
// function will be triggered
type EventHandler func(event Event, c *Client) error

const (
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
	// EventNewMessage is a response to send_message
	EventNewMessage = "new_message"
	// EventChangeRoom is event when switching rooms
	EventChangeRoom = "change_room"
)

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// NewDispatchEvent is returned when responding to send_message
type NewDispatchEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

func SendMessageHandler(event Event, c *Client) error {

	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// process to send message to other clients
	var dispatchevent NewDispatchEvent
	dispatchevent.From = chatevent.From
	dispatchevent.Message = chatevent.Message
	dispatchevent.Sent = time.Now()

	_, err := json.Marshal(dispatchevent)
	if err != nil {
		return fmt.Errorf("failed to marshal dispatchevent message: %v", err)

	}

	return nil
}
