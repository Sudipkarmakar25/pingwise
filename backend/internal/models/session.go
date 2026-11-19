package models

import "time"

// Session stores only the hashes of the tokens handed to the client
type Session struct {
	Id               string    `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	UserId           string    `json:"user_id"`
	OrgId            string    `json:"org_id"`
	AccessTokenHash  string    `json:"-"`
	RefreshTokenHash string    `json:"-"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

func (Session) TableName() string { return "sessions" }
