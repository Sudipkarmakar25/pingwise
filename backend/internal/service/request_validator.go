package service

import (
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

func getJSONFieldName(structType reflect.Type, fieldName string) string {
	if field, found := structType.FieldByName(fieldName); found {
		// body fields use json, path params like :id use param, query params use query
		for _, tag := range []string{"json", "param", "query"} {
			if name := field.Tag.Get(tag); name != "" {
				return name
			}
		}
	}
	return fieldName
}

func FormatValidationErrors(err error, model interface{}) []ValidationErrorResponse {
	var errors []ValidationErrorResponse
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		structType := reflect.TypeOf(model)
		for _, vErr := range validationErrors {
			jsonFieldName := getJSONFieldName(structType, vErr.StructField())
			errors = append(errors, ValidationErrorResponse{
				Field:   jsonFieldName,
				Message: vErr.Tag() + " validation failed",
			})
		}
	}
	return errors
}

func NewValidationError(err error, model interface{}) *Error {
	return &Error{
		Message:          InvalidPayload,
		StatusCode:       http.StatusBadRequest,
		ValidationErrors: FormatValidationErrors(err, model),
	}
}
