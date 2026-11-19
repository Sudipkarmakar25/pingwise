package auth_service

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

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

const invalidCredentialsMessage = "invalid email or password"

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

type Api interface {
	Signup(ctx custom.Context, req *SignUpRequest) (*LoginResponse, *service.Error)
	Login(ctx custom.Context, req *LoginRequest) (*LoginResponse, *service.Error)
	Logout(ctx custom.Context) *service.Error
	ChangePassword(ctx custom.Context, req *ChangePasswordRequest) *service.Error
	ValidateAccessToken(ctx custom.Context, accessToken string) (*custom.UserProfileCtx, *service.Error)
}

type AuthService struct {
	orgRepo     *org_repo.OrgRepo
	userRepo    *user_repo.UserRepo
	sessionRepo *session_repo.SessionRepo
}

var (
	authService     *AuthService
	authServiceOnce sync.Once
)

func GetAuthService() Api {
	authServiceOnce.Do(func() {
		authService = &AuthService{
			orgRepo:     org_repo.GetOrgRepo(),
			userRepo:    user_repo.GetUserRepo(),
			sessionRepo: session_repo.GetSessionRepo(),
		}
	})
	return authService
}

// Signup creates the organisation with the signing up user as its owner
func (s *AuthService) Signup(ctx custom.Context, req *SignUpRequest) (*LoginResponse, *service.Error) {
	email := normaliseEmail(req.Email)
	_, repoErr := s.userRepo.GetByEmail(ctx, email)
	if repoErr == nil {
		return nil, service.NewError("user with this email already exists", http.StatusConflict)
	} else if repoErr.Type != repo.NotFound {
		return nil, service.HandleRepoErr(repoErr)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to hash password : %v", err)
		return nil, service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}

	slug, serviceErr := s.generateOrgSlug(ctx, req.OrgName)
	if serviceErr != nil {
		return nil, serviceErr
	}

	org := &models.Organisation{Name: strings.TrimSpace(req.OrgName), Slug: slug}
	user := &models.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Name:         strings.TrimSpace(req.Name),
		Role:         enums.UserRoleOwner,
	}
	txErr := database.WithTransaction(ctx, func() error {
		if repoErr := s.orgRepo.Create(ctx, org); repoErr != nil {
			return repoErr
		}
		user.OrgId = org.Id
		if repoErr := s.userRepo.Create(ctx, user); repoErr != nil {
			return repoErr
		}
		return nil
	})
	if txErr != nil {
		return nil, service.HandleErr(txErr)
	}

	return s.createSession(ctx, user)
}

func (s *AuthService) Login(ctx custom.Context, req *LoginRequest) (*LoginResponse, *service.Error) {
	user, repoErr := s.userRepo.GetByEmail(ctx, normaliseEmail(req.Email))
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, service.NewError(invalidCredentialsMessage, http.StatusUnauthorized)
		}
		return nil, service.HandleRepoErr(repoErr)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, service.NewError(invalidCredentialsMessage, http.StatusUnauthorized)
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Logout(ctx custom.Context) *service.Error {
	if repoErr := s.sessionRepo.DeleteById(ctx, ctx.GetUser().SessionId); repoErr != nil {
		return service.HandleRepoErr(repoErr)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx custom.Context, req *ChangePasswordRequest) *service.Error {
	user, repoErr := s.userRepo.GetById(ctx, ctx.GetUserId())
	if repoErr != nil {
		return service.HandleRepoErr(repoErr)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return service.NewError("old password is incorrect", http.StatusBadRequest)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to hash password : %v", err)
		return service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}
	if repoErr := s.userRepo.UpdatePassword(ctx, user.Id, string(passwordHash)); repoErr != nil {
		return service.HandleRepoErr(repoErr)
	}
	return nil
}

// ValidateAccessToken resolves an access token to its user. Role and org are read from
// the user row on every request, so a role change or removal applies immediately.
func (s *AuthService) ValidateAccessToken(ctx custom.Context, accessToken string) (*custom.UserProfileCtx, *service.Error) {
	session, repoErr := s.sessionRepo.GetByAccessTokenHash(ctx, hashToken(accessToken))
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, service.NewError("invalid token", http.StatusUnauthorized)
		}
		return nil, service.HandleRepoErr(repoErr)
	}
	if session.AccessExpiresAt.Before(time.Now()) {
		return nil, service.NewError("token expired", http.StatusUnauthorized)
	}

	user, repoErr := s.userRepo.GetById(ctx, session.UserId)
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, service.NewError("invalid user", http.StatusUnauthorized)
		}
		return nil, service.HandleRepoErr(repoErr)
	}

	return &custom.UserProfileCtx{
		UserID:    user.Id,
		OrgId:     user.OrgId,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		SessionId: session.Id,
	}, nil
}

func (s *AuthService) createSession(ctx custom.Context, user *models.User) (*LoginResponse, *service.Error) {
	accessToken, err := generateToken()
	if err != nil {
		logger.Errorf("failed to generate access token : %v", err)
		return nil, service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}
	refreshToken, err := generateToken()
	if err != nil {
		logger.Errorf("failed to generate refresh token : %v", err)
		return nil, service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}

	now := time.Now()
	session := &models.Session{
		UserId:           user.Id,
		OrgId:            user.OrgId,
		AccessTokenHash:  hashToken(accessToken),
		RefreshTokenHash: hashToken(refreshToken),
		AccessExpiresAt:  now.Add(accessTokenExpiry),
		RefreshExpiresAt: now.Add(refreshTokenExpiry),
	}
	if repoErr := s.sessionRepo.Create(ctx, session); repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenExpiry.Seconds()),
		User:         user,
	}, nil
}

// generateOrgSlug builds a url friendly slug from the org name, adding a random suffix if it is taken
func (s *AuthService) generateOrgSlug(ctx custom.Context, orgName string) (string, *service.Error) {
	slug := strings.Trim(nonSlugChars.ReplaceAllString(strings.ToLower(orgName), "-"), "-")
	if slug == "" {
		slug = "org"
	}
	exists, repoErr := s.orgRepo.ExistsBySlug(ctx, slug)
	if repoErr != nil {
		return "", service.HandleRepoErr(repoErr)
	}
	if exists {
		suffix, err := generateToken()
		if err != nil {
			logger.Errorf("failed to generate slug suffix : %v", err)
			return "", service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
		}
		slug = slug + "-" + suffix[:6]
	}
	return slug, nil
}

func normaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
