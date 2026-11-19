package models

import (
	"time"

	"gorm.io/gorm"

	"pingwise/internal/models/enums"
)

type User struct {
	Id                 string         `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	OrgId              string         `json:"org_id"`
	Email              string         `json:"email"`
	PasswordHash       string         `json:"-"`
	Name               string         `json:"name"`
	Role               enums.UserRole `json:"role"`
	AddedBy            *string        `json:"added_by"`
	MustChangePassword bool           `json:"must_change_password"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-"`
}

func (User) TableName() string { return "users" }
