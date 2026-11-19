package user_service

import "pingwise/internal/models"

type MeResponse struct {
	User         *models.User         `json:"user"`
	Organisation *models.Organisation `json:"organisation"`
}
