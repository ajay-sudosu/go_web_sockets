package server

import (
	"log"

	"github.com/labstack/echo/v4"
)

func LaunchServer(e *echo.Echo) {
	if err := e.Start("0.0.0.0:5000"); err != nil {
		log.Fatal(">>> Server exited:", err.Error())
	}
}
