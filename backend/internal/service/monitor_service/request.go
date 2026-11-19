package monitor_service

type CreateServiceRequest struct {
	Name             string            `json:"name" validate:"required,max=100"`
	Url              string            `json:"url" validate:"required,http_url,max=2000"`
	Method           string            `json:"method" validate:"omitempty,oneof=GET POST HEAD"`
	AuthHeaderName   string            `json:"auth_header_name" validate:"omitempty,max=100"`
	AuthToken        string            `json:"auth_token" validate:"omitempty,max=4000"`
	CustomHeaders    map[string]string `json:"custom_headers" validate:"omitempty,max=20,dive,keys,min=1,max=100,endkeys,max=2000"`
	RequestBody      string            `json:"request_body" validate:"omitempty,max=10000"`
	ExpectedStatus   int               `json:"expected_status" validate:"omitempty,min=100,max=599"`
	IntervalMinutes  int               `json:"interval_minutes" validate:"omitempty,oneof=1 2 3 5 10 15 30 60"`
	TimeoutSeconds   int               `json:"timeout_seconds" validate:"omitempty,min=1,max=30"`
	FailureThreshold int               `json:"failure_threshold" validate:"omitempty,min=1,max=5"`
}

// UpdateServiceRequest changes only the fields that are sent.
// auth_token: omit to keep the current token, "" to remove auth, a value to replace it.
// custom_headers: omit to keep, {} to clear, a map to replace.
type UpdateServiceRequest struct {
	Id               string            `param:"id" validate:"required,uuid"`
	Name             *string           `json:"name" validate:"omitnil,min=1,max=100"`
	Url              *string           `json:"url" validate:"omitnil,http_url,max=2000"`
	Method           *string           `json:"method" validate:"omitnil,oneof=GET POST HEAD"`
	AuthHeaderName   *string           `json:"auth_header_name" validate:"omitnil,max=100"`
	AuthToken        *string           `json:"auth_token" validate:"omitnil,max=4000"`
	CustomHeaders    map[string]string `json:"custom_headers" validate:"omitempty,max=20,dive,keys,min=1,max=100,endkeys,max=2000"`
	RequestBody      *string           `json:"request_body" validate:"omitnil,max=10000"`
	ExpectedStatus   *int              `json:"expected_status" validate:"omitnil,min=100,max=599"`
	IntervalMinutes  *int              `json:"interval_minutes" validate:"omitnil,oneof=1 2 3 5 10 15 30 60"`
	TimeoutSeconds   *int              `json:"timeout_seconds" validate:"omitnil,min=1,max=30"`
	FailureThreshold *int              `json:"failure_threshold" validate:"omitnil,min=1,max=5"`
	IsActive         *bool             `json:"is_active"`
}

type ServiceIdRequest struct {
	Id string `param:"id" validate:"required,uuid"`
}

// TimelineRequest asks for a service's uptime and response times grouped into time slots.
// granularity: minute (last 3 hours), five_minutes (last 24 hours) or day (last 90 UTC days).
type TimelineRequest struct {
	Id          string `param:"id" validate:"required,uuid"`
	Granularity string `query:"granularity" validate:"required,oneof=minute five_minutes day"`
}

// ListChecksRequest pages through a service's checks, newest first.
// cursor is the next_cursor of the previous page.
type ListChecksRequest struct {
	Id     string `param:"id" validate:"required,uuid"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=200"`
	Cursor string `query:"cursor"`
}
