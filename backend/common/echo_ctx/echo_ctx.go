package echo_ctx

import (
	"github.com/labstack/echo/v4"

	"pingwise/common/custom"
)

const (
	CustomCtx = "custom_ctx"
)

func SetCustomCtx(echoCtx echo.Context, customCtx custom.Context) {
	echoCtx.Set(CustomCtx, customCtx)
	echoCtx.SetRequest(echoCtx.Request().WithContext(customCtx))
}

func GetCustomCtx(echoCtx echo.Context) custom.Context {
	return echoCtx.Get(CustomCtx).(custom.Context)
}
