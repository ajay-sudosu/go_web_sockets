package handlers

import (
	"abc/chat"
	logger "abc/log"
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/labstack/echo/v4"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type poolAdapter struct {
	pool *sync.Pool
}

func (p *poolAdapter) Get() any {
	a := p.pool.Get().(*bytes.Buffer)
	fmt.Println(a)
	return p.pool.Get().(*bytes.Buffer)
}

func (p *poolAdapter) Put(buf any) {
	p.pool.Put(buf)
}

var chatManager = chat.NewChatManager()

func SocketHandler(c echo.Context) error {
	l := logger.Logger(c)
	l.Info("Request rx for chat")

	ws, err := chat.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	userID := c.QueryParam("user_id")
	if err != nil {
		log.Println("Websocket could not be established as request upgrade failed:", err)
		return err
	}

	chatManager.HandleChat(l, ws, userID)
	return nil
}
