package repo

import (
	"sync"

	"gorm.io/gorm"

	"pingwise/common/custom"
	"pingwise/common/database"
)

type BaseRepo struct{}

var (
	baseRepo     *BaseRepo
	baseRepoOnce sync.Once
)

func GetBaseRepo() *BaseRepo {
	baseRepoOnce.Do(func() {
		baseRepo = &BaseRepo{}
	})
	return baseRepo
}

// GetDb returns the transaction set on ctx if there is one, else the master db
func (r *BaseRepo) GetDb(ctx custom.Context) *gorm.DB {
	if tx := ctx.GetDbTxn(); tx != nil {
		return tx
	}
	return database.GetMasterDb().WithContext(ctx)
}
