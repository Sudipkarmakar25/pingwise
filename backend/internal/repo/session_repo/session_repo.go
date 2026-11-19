package session_repo

import (
	"sync"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/repo"
)

type SessionRepo struct {
	*repo.BaseRepo
}

var (
	sessionRepo     *SessionRepo
	sessionRepoOnce sync.Once
)

func GetSessionRepo() *SessionRepo {
	sessionRepoOnce.Do(func() {
		sessionRepo = &SessionRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return sessionRepo
}

func (r *SessionRepo) Create(ctx custom.Context, session *models.Session) *repo.RepoErr {
	if err := r.GetDb(ctx).Create(session).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *SessionRepo) GetByAccessTokenHash(ctx custom.Context, accessTokenHash string) (*models.Session, *repo.RepoErr) {
	var session models.Session
	if err := r.GetDb(ctx).Where("access_token_hash = ?", accessTokenHash).First(&session).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &session, nil
}

func (r *SessionRepo) DeleteByUserId(ctx custom.Context, userId string) *repo.RepoErr {
	if err := r.GetDb(ctx).Where("user_id = ?", userId).Delete(&models.Session{}).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *SessionRepo) DeleteById(ctx custom.Context, id string) *repo.RepoErr {
	if err := r.GetDb(ctx).Where("id = ?", id).Delete(&models.Session{}).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
