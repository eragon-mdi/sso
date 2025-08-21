package authservice

import "github.com/eragon-mdi/sso/internal/common/configs"

type Auth struct {
	r           Repository
	passHasher  PasswordHasher
	tokener     Tokener
	tokenHasher TokenHasher
	cfg         *configs.BussinesLogic
}

func New(r Repository, ph PasswordHasher, t Tokener, th TokenHasher, c *configs.BussinesLogic) *Auth {
	return &Auth{
		r:           r,
		passHasher:  ph,
		tokener:     t,
		tokenHasher: th,
		cfg:         c,
	}
}
