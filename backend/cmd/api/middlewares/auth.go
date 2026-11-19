package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/service"
	"pingwise/internal/service/auth_service"
)

// AuthTokenMiddleware validates the bearer access token and sets the user on the custom ctx
func AuthTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		customCtx := echo_ctx.GetCustomCtx(c)
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, response.NewErrorResponse(
				service.NewError("Unauthorized", http.StatusUnauthorized)))
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, response.NewErrorResponse(
				service.NewError("missing token", http.StatusUnauthorized)))
		}

		userProfile, serviceErr := auth_service.GetAuthService().ValidateAccessToken(customCtx, tokenString)
		if serviceErr != nil {
			return c.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
		}
		customCtx.SetUser(userProfile)
		return next(c)
	}
}
