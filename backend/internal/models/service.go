package models

import (
	"time"

	"gorm.io/gorm"

	"pingwise/internal/models/enums"
)

// Service is an endpoint an organisation wants checked. The auth token is stored only
// encrypted and is never serialised.
type Service struct {
	Id                  string              `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	OrgId               string              `json:"org_id"`
	Name                string              `json:"name"`
	Url                 string              `json:"url"`
	Method              string              `json:"method"`
	AuthHeaderName      *string             `json:"auth_header_name"`
	AuthTokenEncrypted  []byte              `json:"-"`
	AuthTokenLast4      *string             `json:"-"`
	AuthKeyVersion      *int                `json:"-"`
	CustomHeaders       map[string]string   `json:"custom_headers" gorm:"serializer:json"`
	RequestBody         *string             `json:"request_body"`
	ExpectedStatus      int                 `json:"expected_status"`
	IntervalMinutes     int                 `json:"interval_minutes"`
	TimeoutSeconds      int                 `json:"timeout_seconds"`
	FailureThreshold    int                 `json:"failure_threshold"`
	ConsecutiveFailures int                 `json:"consecutive_failures"`
	IsActive            bool                `json:"is_active"`
	CurrentStatus       enums.ServiceStatus `json:"current_status"`
	LastScheduledAt     *time.Time          `json:"-"`
	LastCheckedAt       *time.Time          `json:"last_checked_at"`
	LastResponseTimeMs  *int                `json:"last_response_time_ms"`
	CreatedBy           string              `json:"created_by"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	DeletedAt           gorm.DeletedAt      `json:"-"`
}

func (Service) TableName() string { return "services" }
