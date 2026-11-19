package service

import (
	"fmt"
	"net/http"

	"pingwise/internal/repo"
)

const (
	SomethingUnExpectedErr = "unexpected error. please try again"
	InvalidPayload         = "Invalid request payload"
)

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Message          string      `json:"message"`
	StatusCode       int         `json:"status_code"`
	ErrorCode        string      `json:"error_code"`
	ValidationErrors interface{} `json:"validation_errors"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Message: %s, StatusCode: %d, ErrorCode: %v", e.Message, e.StatusCode, e.ErrorCode)
}

func NewError(msg string, statusCode int) *Error {
	return &Error{Message: msg, StatusCode: statusCode}
}

func HandleRepoErr(err *repo.RepoErr) *Error {
	if err.Type == repo.NotFound {
		return &Error{Message: err.Message, StatusCode: http.StatusNotFound}
	} else if err.Type == repo.InvalidRequest {
		return &Error{Message: err.Message, StatusCode: http.StatusBadRequest}
	} else if err.Type == repo.UniqueKeyViolation {
		return &Error{Message: err.Message, StatusCode: http.StatusConflict}
	}
	return &Error{Message: SomethingUnExpectedErr, StatusCode: http.StatusInternalServerError}
}

func HandleErr(err error) *Error {
	// Check if error is already *Error type
	if e, ok := err.(*Error); ok {
		return e
	}

	// Check if error is *repo.RepoErr type
	if repoErr, ok := err.(*repo.RepoErr); ok {
		return HandleRepoErr(repoErr)
	}

	return &Error{Message: SomethingUnExpectedErr, StatusCode: http.StatusInternalServerError}
}
