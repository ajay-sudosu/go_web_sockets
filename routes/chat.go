package routes

import (
	"abc/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	chtGrp := e.Group("/api/v1")

	chtGrp.GET("/ws", handlers.SocketHandler)

}
