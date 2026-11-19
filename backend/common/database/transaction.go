package database

import (
	"gorm.io/gorm"

	"pingwise/common/custom"
)

// WithTransaction runs fn inside a transaction. Repos pick the transaction up from ctx,
// so every repo call made with this ctx inside fn is part of it.
func WithTransaction(ctx custom.Context, fn func() error) error {
	return GetMasterDb().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx.SetDbTxn(tx)
		defer ctx.SetDbTxn(nil)
		return fn()
	})
}
