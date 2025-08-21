package sqlrepo

import (
	"context"

	"github.com/go-faster/errors"
)

func (r sqlRepo) CheckUserIsAdminByID(ctx context.Context, userID string) (bool, error) {
	var ok bool
	if err := r.s.QueryRowContext(ctx, queryCheckUserIsAdmin, userID).Scan(&ok); err != nil {
		return false, errors.Wrap(err, ErrFailedScan)
	}

	return ok, nil
}
