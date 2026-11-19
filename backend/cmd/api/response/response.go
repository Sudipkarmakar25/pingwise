package response

import (
	"github.com/labstack/echo/v4"
)

type BaseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func NewSuccessfulResponse(ctx echo.Context, data interface{}, statusCode int) error {
	return ctx.JSON(statusCode, &BaseResponse{
		Success: true,
		Data:    data,
	})
}

func NewErrorResponse(data interface{}) *BaseResponse {
	return &BaseResponse{
		Success: false,
		Data:    data,
	}
}
