package sqlrepo

import (
	"context"
	"database/sql"

	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/go-faster/errors"
	"github.com/lib/pq"
)

func (r sqlRepo) NewUser(ctx context.Context, u domain.User) (domain.User, error) {
	row := r.s.QueryRowContext(ctx, queryInsertUser, u.ID, u.Email, u.Password)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.Password); err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, errors.Wrap(domain.ErrDuplicate, ErrFailedExec)
		}
		return domain.User{}, errors.Wrap(err, ErrFailedScan)
	}

	return user, nil
}

func (r sqlRepo) GetUserInfoByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.s.QueryRowContext(ctx, queryGetUserByEmail, email)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Wrap(domain.ErrNotFound, ErrFailedQuery)
		}
		return domain.User{}, errors.Wrap(err, ErrFailedScan)
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505" // unique_violation
	}
	return false
}
