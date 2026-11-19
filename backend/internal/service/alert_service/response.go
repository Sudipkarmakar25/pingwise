package alert_service

import (
	"time"

	"pingwise/internal/models/enums"
)

// MemberAlertResponse is one member's alert setting for a service, as seen by the caller
type MemberAlertResponse struct {
	UserId        string         `json:"user_id"`
	Name          string         `json:"name"`
	Role          enums.UserRole `json:"role"`
	Enabled       bool           `json:"enabled"`
	CanChange     bool           `json:"can_change"` // whether the caller may turn it on or off
	UpdatedByName *string        `json:"updated_by_name"`
	UpdatedAt     *time.Time     `json:"updated_at"`
}
