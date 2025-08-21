package service

import (
	"github.com/eragon-mdi/sso/internal/common/configs"
	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
	"github.com/eragon-mdi/sso/internal/service/sso/auth/hasher"
	hashertokener "github.com/eragon-mdi/sso/internal/service/sso/auth/hasher-tokener"
	tokener "github.com/eragon-mdi/sso/internal/service/sso/auth/tokener"
	permissionservice "github.com/eragon-mdi/sso/internal/service/sso/permission"
	"github.com/eragon-mdi/sso/internal/transport"
	"github.com/go-faster/errors"
)

type service struct {
	r Repository
	sso
}

func New(r Repository, cfg *configs.BussinesLogic) (transport.Service, error) {
	t, err := tokener.New(cfg.PathSecretPrivate, cfg.PathSecretPublic)
	if err != nil {
		return nil, errors.Wrap(err, "failed init tokener")
	}

	return &service{
		r: r,
		sso: sso{
			Auth: authservice.New(
				r,
				hasher.New(cfg.PassHasherCost),
				t,
				hashertokener.New([]byte(cfg.SecretForTokerHasher)),
				cfg),

			Permission: permissionservice.New(r),
		},
	}, nil
}

type Repository interface {
	authservice.Repository
	permissionservice.Repository
}

type sso struct {
	*authservice.Auth
	*permissionservice.Permission
}
