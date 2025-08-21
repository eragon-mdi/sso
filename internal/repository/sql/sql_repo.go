package sqlrepo

import (
	sqlstore "github.com/eragon-mdi/go-playground/storage/sql"
	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
	permissionservice "github.com/eragon-mdi/sso/internal/service/sso/permission"
)

type SqlRepo interface {
	authservice.UserRepository
	permissionservice.UserRepository
}

type sqlRepo struct {
	s sqlstore.Storage
}

func New(s sqlstore.Storage) SqlRepo {
	return &sqlRepo{
		s: s,
	}
}

const (
	ErrFailedQuery        = "repo: failed query"
	ErrFailedExec         = "repo: failed exec"
	ErrFailedScan         = "repo: failed to scan row"
	ErrFailedAffectedRows = "repo: failed to get number of affected rows"
	ErrFailedStartTX      = "repo: failed to start tx"
	ErrFailedCommitTX     = "repo: failed to commit tx"
	ErrFailedRollbackTX   = "repo: failed rollback tx"
	ErrRowsIterations     = "repo: rows iteration error"
)
