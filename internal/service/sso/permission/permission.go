package permissionservice

import (
	"context"

	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/go-faster/errors"
)

//go:generate mockery --name=UserRepository --with-expecter --output=./mocks/userrepo --exported
type UserRepository interface {
	CheckUserIsAdminByID(context.Context, string) (bool, error)
}

func (s Permission) IsAdmin(ctx context.Context, u domain.User) (bool, error) {
	isAdmin, err := s.r.CheckUserIsAdminByID(ctx, u.ID)
	if err != nil {
		return false, errors.Wrap(err, "failed check user for admin privilages")
	}

	return isAdmin, nil
}
