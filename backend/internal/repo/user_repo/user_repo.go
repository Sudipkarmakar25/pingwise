package user_repo

import (
	"sync"

	"gorm.io/gorm/clause"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
)

type UserRepo struct {
	*repo.BaseRepo
}

var (
	userRepo     *UserRepo
	userRepoOnce sync.Once
)

func GetUserRepo() *UserRepo {
	userRepoOnce.Do(func() {
		userRepo = &UserRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return userRepo
}

func (r *UserRepo) Create(ctx custom.Context, user *models.User) *repo.RepoErr {
	if err := r.GetDb(ctx).Create(user).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *UserRepo) GetById(ctx custom.Context, id string) (*models.User, *repo.RepoErr) {
	var user models.User
	if err := r.GetDb(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &user, nil
}

func (r *UserRepo) GetByIdAndOrgId(ctx custom.Context, id, orgId string) (*models.User, *repo.RepoErr) {
	var user models.User
	if err := r.GetDb(ctx).Where("id = ? AND org_id = ?", id, orgId).First(&user).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx custom.Context, email string) (*models.User, *repo.RepoErr) {
	var user models.User
	if err := r.GetDb(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &user, nil
}

func (r *UserRepo) ListByOrgId(ctx custom.Context, orgId string) ([]models.User, *repo.RepoErr) {
	var users []models.User
	if err := r.GetDb(ctx).Where("org_id = ?", orgId).Order("created_at").Find(&users).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return users, nil
}

// LockOwnerIds locks the org's owner rows until the transaction ends, so two requests
// cannot remove the last owners at the same time. Must be called inside a transaction.
func (r *UserRepo) LockOwnerIds(ctx custom.Context, orgId string) ([]string, *repo.RepoErr) {
	var ids []string
	err := r.GetDb(ctx).Model(&models.User{}).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("org_id = ? AND role = ?", orgId, enums.UserRoleOwner).Pluck("id", &ids).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return ids, nil
}

func (r *UserRepo) UpdateRole(ctx custom.Context, id string, role enums.UserRole) *repo.RepoErr {
	if err := r.GetDb(ctx).Model(&models.User{}).Where("id = ?", id).Update("role", role).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx custom.Context, id, passwordHash string) *repo.RepoErr {
	err := r.GetDb(ctx).Model(&models.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{"password_hash": passwordHash, "must_change_password": false}).Error
	if err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

// SoftDelete sets deleted_at; gorm then hides the row from every other query
func (r *UserRepo) SoftDelete(ctx custom.Context, id string) *repo.RepoErr {
	if err := r.GetDb(ctx).Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
