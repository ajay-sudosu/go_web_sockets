package main

import (
	logger "abc/log"
	middlewares "abc/middleware"
	"abc/routes"
	"abc/server"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	l := logger.SetLogger()

	middlewares.InjectLogger(e, l)

	routes.RegisterRoutes(e)
	server.LaunchServer(e)
}
