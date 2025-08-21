package authservice

import (
	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/go-faster/errors"
)

const (
	ErrInvalidToken       = "invalid token"
	ErrUnauthenticatedCtx = "unauthenticated: ctx mismatch"
	ErrFailedGenJWT       = "failed generate jwt"
	ErrFailedHashRefresh  = "failed hashing refresh token"
)

func (s *Auth) verificationToken(refresh string, userDctx domain.DeviceCtx) (domain.Meta, error) {
	m, err := s.tokener.VerifyRefresh([]byte(refresh))
	if err != nil {
		return domain.Meta{}, errors.Wrap(err, ErrInvalidToken)
	}

	if !userDctx.Compare(m.Ctx) {
		return domain.Meta{}, errors.New(ErrUnauthenticatedCtx)
	}

	return m, nil
}

func (s *Auth) genTokensFlow(userId string, dctx domain.DeviceCtx) (*domain.Token, *domain.RefreshToken, error) {
	m := domain.NewMeta(s.cfg.TokenTTL, userId, dctx.AppId, dctx.DeviceID)

	access, refresh, err := s.tokener.GenPair(m)
	if err != nil {
		return nil, nil, errors.Wrap(err, ErrFailedGenJWT)
	}

	refreshHash, err := s.tokenHasher.Sum([]byte(refresh))
	if err != nil {
		return nil, nil, errors.Wrap(err, ErrFailedHashRefresh)
	}

	token := domain.NewToken(string(access), string(refresh))
	rt := domain.NewRefreshTorken(string(refreshHash), m)

	return &token, &rt, nil
}
