package authservice

import (
	"context"

	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/go-faster/errors"
	"github.com/google/uuid"
)

//go:generate mockery --name=Repository --with-expecter --output=./mocks/repository --exported
type Repository interface {
	UserRepository
	TokenRepository
}

type UserRepository interface {
	NewUser(context.Context, domain.User) (domain.User, error)
	GetUserInfoByEmail(context.Context, string) (domain.User, error)
}

type TokenRepository interface {
	SaveRefreshToken(context.Context, domain.RefreshToken) error
	RotateToken(_ context.Context, oldHash string, newRT domain.RefreshToken) error
	RevokeTokenByHash(context.Context, string) error
}

//go:generate mockery --name=PasswordHasher --with-expecter --output=./mocks/password-hasher --exported
type PasswordHasher interface {
	Gen([]byte) ([]byte, error)
	Compare(hash []byte, pass []byte) (bool, error)
}

//go:generate mockery --name=Tokener --with-expecter --output=./mocks/tokener --exported
type Tokener interface {
	GenPair(domain.Meta) (access, refresh []byte, err error)
	VerifyRefresh([]byte) (domain.Meta, error)
}

//go:generate mockery --name=TokenHasher --with-expecter --output=./mocks/token-hasher --exported
type TokenHasher interface {
	Sum([]byte) ([]byte, error) // например, HMAC-SHA256(secret, token)
}

const (
	ErrFailedHashPass      = "failed to hash pass"
	ErrFailedSaveUser      = "failed save new user in repo"
	ErrDuplicateEmail      = "cause: duplicate email"
	ErrFailedGetUserInfo   = "failed get user info"
	ErrFailedCheckPass     = "failed check pass"
	ErrFailedGenerateToken = "failed generate jwt-pair"
	ErrFailedSaveToken     = "failed save refreshed token"
	ErrFailedVerifyToken   = "failed verifaicate token"
	ErrFailedHashToken     = "failed hashing refresh token"
	ErrFailedRotateToken   = "rotate failed"
	ErrFailedRevokeToken   = "failed get refresh token: internal"
)

func (s *Auth) Register(ctx context.Context, u domain.User) (domain.User, error) {
	hashedPass, err := s.passHasher.Gen([]byte(u.Password))
	if err != nil {
		return domain.User{}, errors.Wrap(err, ErrFailedHashPass)
	}
	u.SetID(uuid.NewString())
	u.SetPass(string(hashedPass))

	user, err := s.r.NewUser(ctx, u)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			return domain.User{}, errors.Wrap(domain.ErrDuplicate, ErrDuplicateEmail)
		}
		return domain.User{}, errors.Wrap(err, ErrFailedSaveUser)
	}

	return user, nil
}

func (s *Auth) Login(ctx context.Context, u domain.User, dctx domain.DeviceCtx) (domain.Token, error) {
	originPass := []byte(u.Password)

	// авторизация
	u, err := s.r.GetUserInfoByEmail(ctx, u.Email)
	if err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedGetUserInfo)
	}
	hashedPass := u.Password
	isCorrect, err := s.passHasher.Compare([]byte(hashedPass), originPass)
	if err != nil || !isCorrect {
		return domain.Token{}, errors.Wrap(err, ErrFailedCheckPass)
	}

	token, newRt, err := s.genTokensFlow(u.ID, dctx)
	if err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedGenerateToken)
	}

	if err := s.r.SaveRefreshToken(ctx, *newRt); err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedSaveToken)
	}

	return *token, nil
}

func (s *Auth) Refresh(ctx context.Context, oldRefresh string, dctx domain.DeviceCtx) (domain.Token, error) {
	m, err := s.verificationToken(oldRefresh, dctx)
	if err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedVerifyToken)
	}

	token, newRt, err := s.genTokensFlow(m.UserID, m.Ctx)
	if err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedGenerateToken)
	}

	oldRefreshHash, err := s.tokenHasher.Sum([]byte(oldRefresh))
	if err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedHashToken)
	}
	if err := s.r.RotateToken(ctx, string(oldRefreshHash), *newRt); err != nil {
		return domain.Token{}, errors.Wrap(err, ErrFailedRotateToken)
	}

	return *token, nil
}

func (s *Auth) Logout(ctx context.Context, refresh string, dctx domain.DeviceCtx) error {
	if _, err := s.verificationToken(refresh, dctx); err != nil {
		return errors.Wrap(err, ErrFailedVerifyToken)
	}

	hashBytes, err := s.tokenHasher.Sum([]byte(refresh))
	if err != nil {
		return errors.Wrap(err, ErrFailedHashToken)
	}

	if err := s.r.RevokeTokenByHash(ctx, string(hashBytes)); err != nil && !errors.Is(err, domain.ErrNotFound) {
		return errors.Wrap(err, ErrFailedRevokeToken)
	}

	return nil
}
