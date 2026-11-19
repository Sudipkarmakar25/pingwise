package middlewares

import (
	"github.com/labstack/echo/v4"

	"pingwise/common/custom"
	"pingwise/common/echo_ctx"
)

func CustomCtxMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		echo_ctx.SetCustomCtx(c, custom.NewCustomContext(c.Request().Context()))
		return next(c)
	}
}
