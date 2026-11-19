package models

import "time"

// AlertSubscription says whether a member gets alert emails for a service. Without a row,
// owners get them and editors and viewers do not.
type AlertSubscription struct {
	Id        string    `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	OrgId     string    `json:"org_id"`
	UserId    string    `json:"user_id"`
	ServiceId string    `json:"service_id"`
	IsEnabled bool      `json:"is_enabled"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AlertSubscription) TableName() string { return "alert_subscriptions" }
