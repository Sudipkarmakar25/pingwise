package alert_service

type ServiceIdRequest struct {
	Id string `param:"id" validate:"required,uuid"`
}

// UpdateAlertRequest turns a member's alert emails for a service on or off
type UpdateAlertRequest struct {
	Id      string `param:"id" validate:"required,uuid"`
	UserId  string `param:"userId" validate:"required,uuid"`
	Enabled *bool  `json:"enabled" validate:"required"`
}
