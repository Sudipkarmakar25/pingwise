package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/service"
	"pingwise/internal/service/auth_service"
)

// AuthHandler struct to handle authentication related routes
type AuthHandler struct {
	authService auth_service.Api
}

var authHandler *AuthHandler

func GetAuthHandler() *AuthHandler {
	if authHandler == nil {
		authHandler = &AuthHandler{authService: auth_service.GetAuthService()}
	}
	return authHandler
}

// Signup creates a new organisation and its owner
func (h *AuthHandler) Signup(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req auth_service.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.authService.Signup(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusCreated)
}

func (h *AuthHandler) Login(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req auth_service.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.authService.Login(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *AuthHandler) Logout(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	if serviceErr := h.authService.Logout(customCtx); serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, map[string]string{"message": "logged out"}, http.StatusOK)
}

func (h *AuthHandler) ChangePassword(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req auth_service.ChangePasswordRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	if serviceErr := h.authService.ChangePassword(customCtx, &req); serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, map[string]string{"message": "password changed"}, http.StatusOK)
}
