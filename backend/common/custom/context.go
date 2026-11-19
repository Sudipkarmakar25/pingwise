package custom

import (
	"context"

	"gorm.io/gorm"

	"pingwise/internal/models/enums"
)

type UserProfileCtx struct {
	UserID    string
	OrgId     string
	Name      string
	Email     string
	Role      enums.UserRole
	SessionId string
}

// Context is our custom context interface that extends context.Context
type Context interface {
	context.Context
	SetUser(userProfile *UserProfileCtx)
	GetUser() *UserProfileCtx
	GetUserId() string
	GetOrgId() string
	SetDbTxn(tx *gorm.DB)
	GetDbTxn() *gorm.DB
}

// customContextImpl is the concrete implementation of Context
type customContextImpl struct {
	context.Context
	user  *UserProfileCtx
	dbTxn *gorm.DB
}

// NewCustomContext creates a new Context
func NewCustomContext(parentCtx context.Context) Context {
	return &customContextImpl{Context: parentCtx}
}

func (c *customContextImpl) SetUser(userProfile *UserProfileCtx) {
	c.user = userProfile
}

func (c *customContextImpl) GetUser() *UserProfileCtx {
	return c.user
}

func (c *customContextImpl) GetUserId() string {
	if c.user == nil {
		return ""
	}
	return c.user.UserID
}

func (c *customContextImpl) GetOrgId() string {
	if c.user == nil {
		return ""
	}
	return c.user.OrgId
}

func (c *customContextImpl) SetDbTxn(tx *gorm.DB) {
	c.dbTxn = tx
}

func (c *customContextImpl) GetDbTxn() *gorm.DB {
	return c.dbTxn
}
