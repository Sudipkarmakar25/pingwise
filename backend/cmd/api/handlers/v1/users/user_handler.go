package users

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/service"
	"pingwise/internal/service/user_service"
)

// UserHandler struct to handle user management routes
type UserHandler struct {
	userService user_service.Api
}

var userHandler *UserHandler

func GetUserHandler() *UserHandler {
	if userHandler == nil {
		userHandler = &UserHandler{userService: user_service.GetUserService()}
	}
	return userHandler
}

func (h *UserHandler) GetMe(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	res, serviceErr := h.userService.GetMe(customCtx)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *UserHandler) ListUsers(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	res, serviceErr := h.userService.ListUsers(customCtx)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *UserHandler) AddUser(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req user_service.AddUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.userService.AddUser(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusCreated)
}

func (h *UserHandler) UpdateUserRole(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req user_service.UpdateUserRoleRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.userService.UpdateUserRole(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *UserHandler) RemoveUser(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req user_service.UserIdRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	if serviceErr := h.userService.RemoveUser(customCtx, &req); serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, map[string]string{"message": "user removed"}, http.StatusOK)
}
