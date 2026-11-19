package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/models/enums"
	"pingwise/internal/service"
)

// RoleMiddleware allows the request only if the user's role is one of allowedRoles.
// It must run after AuthTokenMiddleware.
func RoleMiddleware(allowedRoles ...enums.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := echo_ctx.GetCustomCtx(c).GetUser()
			if user != nil {
				for _, role := range allowedRoles {
					if user.Role == role {
						return next(c)
					}
				}
			}
			return c.JSON(http.StatusForbidden, response.NewErrorResponse(
				service.NewError("you don't have permission to perform this action", http.StatusForbidden)))
		}
	}
}
