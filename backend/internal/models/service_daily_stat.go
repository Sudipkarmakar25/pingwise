package models

import "time"

// ServiceDailyStat is one row per service per UTC day, used for the 90 day uptime bars
type ServiceDailyStat struct {
	ServiceId       string    `json:"service_id" gorm:"primaryKey"`
	OrgId           string    `json:"org_id"`
	Day             time.Time `json:"day" gorm:"primaryKey;type:date"`
	TotalChecks     int       `json:"total_checks"`
	UpChecks        int       `json:"up_checks"`
	TotalResponseMs int64     `json:"total_response_ms"`
	MaxResponseMs   int       `json:"max_response_ms"`
}

func (ServiceDailyStat) TableName() string { return "service_daily_stats" }
