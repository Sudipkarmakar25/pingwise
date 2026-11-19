package org_repo

import (
	"sync"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/repo"
)

type OrgRepo struct {
	*repo.BaseRepo
}

var (
	orgRepo     *OrgRepo
	orgRepoOnce sync.Once
)

func GetOrgRepo() *OrgRepo {
	orgRepoOnce.Do(func() {
		orgRepo = &OrgRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return orgRepo
}

func (r *OrgRepo) Create(ctx custom.Context, org *models.Organisation) *repo.RepoErr {
	if err := r.GetDb(ctx).Create(org).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *OrgRepo) GetById(ctx custom.Context, id string) (*models.Organisation, *repo.RepoErr) {
	var org models.Organisation
	if err := r.GetDb(ctx).Where("id = ?", id).First(&org).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &org, nil
}

func (r *OrgRepo) ExistsBySlug(ctx custom.Context, slug string) (bool, *repo.RepoErr) {
	var count int64
	if err := r.GetDb(ctx).Model(&models.Organisation{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, repo.HandleErr(err)
	}
	return count > 0, nil
}
