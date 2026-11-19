package services

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/service"
	"pingwise/internal/service/monitor_service"
)

// ServiceHandler struct to handle monitored service routes
type ServiceHandler struct {
	monitorService monitor_service.Api
}

var serviceHandler *ServiceHandler

func GetServiceHandler() *ServiceHandler {
	if serviceHandler == nil {
		serviceHandler = &ServiceHandler{monitorService: monitor_service.GetMonitorService()}
	}
	return serviceHandler
}

func (h *ServiceHandler) CreateService(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.CreateServiceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.monitorService.CreateService(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusCreated)
}

// ListServices returns every service of the org with its current status. The dashboard's Refresh button calls this.
func (h *ServiceHandler) ListServices(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	res, serviceErr := h.monitorService.ListServices(customCtx)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *ServiceHandler) GetService(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.ServiceIdRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.monitorService.GetService(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

// ListStats returns the daily stats of every service in the org, for the dashboard's uptime bars
func (h *ServiceHandler) ListStats(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	res, serviceErr := h.monitorService.ListStats(customCtx)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

// GetTimeline returns a service's uptime and response times grouped by minute, 5 minutes or day
func (h *ServiceHandler) GetTimeline(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.TimelineRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.monitorService.GetTimeline(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

// ListChecks returns a service's check history, newest first, one page at a time
func (h *ServiceHandler) ListChecks(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.ListChecksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.monitorService.ListChecks(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *ServiceHandler) UpdateService(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.UpdateServiceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.monitorService.UpdateService(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

func (h *ServiceHandler) DeleteService(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req monitor_service.ServiceIdRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	if serviceErr := h.monitorService.DeleteService(customCtx, &req); serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, map[string]string{"message": "service deleted"}, http.StatusOK)
}
