package middlewares

import (
	logging "abc/log"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Middleware to inject the logger into the context context
func InjectLogger(e *echo.Echo, log *zap.Logger) {
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := logging.AddContext(c.Request().Context(), log)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
}
