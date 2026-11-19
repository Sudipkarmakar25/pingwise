package models

import "time"

type Organisation struct {
	Id        string    `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Organisation) TableName() string { return "organisations" }
