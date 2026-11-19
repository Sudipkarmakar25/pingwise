package user_service

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"pingwise/common/custom"
	"pingwise/common/database"
	"pingwise/common/logger"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
	"pingwise/internal/repo/org_repo"
	"pingwise/internal/repo/session_repo"
	"pingwise/internal/repo/user_repo"
	"pingwise/internal/service"
)

const lastOwnerMessage = "organisation must have at least one owner"

type Api interface {
	GetMe(ctx custom.Context) (*MeResponse, *service.Error)
	ListUsers(ctx custom.Context) ([]models.User, *service.Error)
	AddUser(ctx custom.Context, req *AddUserRequest) (*models.User, *service.Error)
	UpdateUserRole(ctx custom.Context, req *UpdateUserRoleRequest) (*models.User, *service.Error)
	RemoveUser(ctx custom.Context, req *UserIdRequest) *service.Error
}

type UserService struct {
	orgRepo     *org_repo.OrgRepo
	userRepo    *user_repo.UserRepo
	sessionRepo *session_repo.SessionRepo
}

var (
	userService     *UserService
	userServiceOnce sync.Once
)

func GetUserService() Api {
	userServiceOnce.Do(func() {
		userService = &UserService{
			orgRepo:     org_repo.GetOrgRepo(),
			userRepo:    user_repo.GetUserRepo(),
			sessionRepo: session_repo.GetSessionRepo(),
		}
	})
	return userService
}

func (s *UserService) GetMe(ctx custom.Context) (*MeResponse, *service.Error) {
	user, repoErr := s.userRepo.GetById(ctx, ctx.GetUserId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	org, repoErr := s.orgRepo.GetById(ctx, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	return &MeResponse{User: user, Organisation: org}, nil
}

func (s *UserService) ListUsers(ctx custom.Context) ([]models.User, *service.Error) {
	users, repoErr := s.userRepo.ListByOrgId(ctx, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	return users, nil
}

// AddUser creates a user in the caller's org with a temporary password they must change
func (s *UserService) AddUser(ctx custom.Context, req *AddUserRequest) (*models.User, *service.Error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	_, repoErr := s.userRepo.GetByEmail(ctx, email)
	if repoErr == nil {
		return nil, service.NewError("user with this email already exists", http.StatusConflict)
	} else if repoErr.Type != repo.NotFound {
		return nil, service.HandleRepoErr(repoErr)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.TemporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to hash password : %v", err)
		return nil, service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}

	addedBy := ctx.GetUserId()
	user := &models.User{
		OrgId:              ctx.GetOrgId(),
		Email:              email,
		PasswordHash:       string(passwordHash),
		Name:               strings.TrimSpace(req.Name),
		Role:               req.Role,
		AddedBy:            &addedBy,
		MustChangePassword: true,
	}
	if repoErr := s.userRepo.Create(ctx, user); repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	return user, nil
}

func (s *UserService) UpdateUserRole(ctx custom.Context, req *UpdateUserRoleRequest) (*models.User, *service.Error) {
	var user *models.User
	txErr := database.WithTransaction(ctx, func() error {
		ownerIds, repoErr := s.userRepo.LockOwnerIds(ctx, ctx.GetOrgId())
		if repoErr != nil {
			return repoErr
		}
		user, repoErr = s.userRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
		if repoErr != nil {
			return repoErr
		}
		if user.Role == enums.UserRoleOwner && req.Role != enums.UserRoleOwner && len(ownerIds) <= 1 {
			return service.NewError(lastOwnerMessage, http.StatusBadRequest)
		}
		if repoErr := s.userRepo.UpdateRole(ctx, user.Id, req.Role); repoErr != nil {
			return repoErr
		}
		user.Role = req.Role
		return nil
	})
	if txErr != nil {
		return nil, service.HandleErr(txErr)
	}
	return user, nil
}

// RemoveUser soft deletes the user and ends all their sessions
func (s *UserService) RemoveUser(ctx custom.Context, req *UserIdRequest) *service.Error {
	txErr := database.WithTransaction(ctx, func() error {
		ownerIds, repoErr := s.userRepo.LockOwnerIds(ctx, ctx.GetOrgId())
		if repoErr != nil {
			return repoErr
		}
		user, repoErr := s.userRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
		if repoErr != nil {
			return repoErr
		}
		if user.Role == enums.UserRoleOwner && len(ownerIds) <= 1 {
			return service.NewError(lastOwnerMessage, http.StatusBadRequest)
		}
		if repoErr := s.userRepo.SoftDelete(ctx, user.Id); repoErr != nil {
			return repoErr
		}
		if repoErr := s.sessionRepo.DeleteByUserId(ctx, user.Id); repoErr != nil {
			return repoErr
		}
		return nil
	})
	if txErr != nil {
		return service.HandleErr(txErr)
	}
	return nil
}
