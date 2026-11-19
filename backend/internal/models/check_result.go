package models

import "time"

type CheckResult struct {
	Id             int64     `json:"id" gorm:"primaryKey"`
	ServiceId      string    `json:"service_id"`
	OrgId          string    `json:"org_id"`
	CheckedAt      time.Time `json:"checked_at"`
	IsUp           bool      `json:"is_up"`
	StatusCode     *int      `json:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	Error          *string   `json:"error"`
}

func (CheckResult) TableName() string { return "check_results" }
