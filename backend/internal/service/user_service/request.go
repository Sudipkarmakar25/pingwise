package user_service

import "pingwise/internal/models/enums"

type AddUserRequest struct {
	Name              string         `json:"name" validate:"required,max=100"`
	Email             string         `json:"email" validate:"required,email"`
	Role              enums.UserRole `json:"role" validate:"required,oneof=owner editor viewer"`
	TemporaryPassword string         `json:"temporary_password" validate:"required,min=8,max=72"`
}

type UpdateUserRoleRequest struct {
	Id   string         `param:"id" validate:"required,uuid"`
	Role enums.UserRole `json:"role" validate:"required,oneof=owner editor viewer"`
}

type UserIdRequest struct {
	Id string `param:"id" validate:"required,uuid"`
}
