package alerts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/response"
	"pingwise/common/echo_ctx"
	"pingwise/internal/service"
	"pingwise/internal/service/alert_service"
)

// AlertHandler struct to handle who gets alert emails for a service
type AlertHandler struct {
	alertService alert_service.Api
}

var alertHandler *AlertHandler

func GetAlertHandler() *AlertHandler {
	if alertHandler == nil {
		alertHandler = &AlertHandler{alertService: alert_service.GetAlertService()}
	}
	return alertHandler
}

// ListServiceAlerts returns every member of the org with whether they get alerts for the service
func (h *AlertHandler) ListServiceAlerts(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req alert_service.ServiceIdRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.alertService.ListServiceAlerts(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}

// UpdateServiceAlert turns a member's alerts for a service on or off
func (h *AlertHandler) UpdateServiceAlert(ctx echo.Context) error {
	customCtx := echo_ctx.GetCustomCtx(ctx)
	var req alert_service.UpdateAlertRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(
			service.NewError(service.InvalidPayload, http.StatusBadRequest)))
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(service.NewValidationError(err, req)))
	}
	res, serviceErr := h.alertService.UpdateServiceAlert(customCtx, &req)
	if serviceErr != nil {
		return ctx.JSON(serviceErr.StatusCode, response.NewErrorResponse(serviceErr))
	}
	return response.NewSuccessfulResponse(ctx, res, http.StatusOK)
}
