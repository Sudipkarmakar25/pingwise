package repo

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"pingwise/common/logger"
)

type ErrType string

const (
	NotFound           ErrType = "not found"
	InvalidRequest     ErrType = "invalid request"
	UniqueKeyViolation ErrType = "unique key violation"
	UnexpectedError    ErrType = "UNEXPECTED_ERROR"
)

const pgUniqueViolationCode = "23505"

type RepoErr struct {
	Message string
	Type    ErrType
}

func (r *RepoErr) Error() string {
	return r.Message
}

func NewRepoErr(Type ErrType, Message string) *RepoErr {
	return &RepoErr{
		Message: Message,
		Type:    Type,
	}
}

// HandleErr converts a gorm/postgres error into a RepoErr
func HandleErr(err error) *RepoErr {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewRepoErr(NotFound, "record not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return NewRepoErr(UniqueKeyViolation, "record already exists")
	}
	logger.Errorf("database error : %v", err)
	return NewRepoErr(UnexpectedError, err.Error())
}
