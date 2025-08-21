package authservice

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/eragon-mdi/sso/internal/common/configs"
	"github.com/eragon-mdi/sso/internal/domain"

	mocks_hasher "github.com/eragon-mdi/sso/internal/service/sso/auth/mocks/password-hasher"
	mocks_repo "github.com/eragon-mdi/sso/internal/service/sso/auth/mocks/repository"
	mocks_tokenhasher "github.com/eragon-mdi/sso/internal/service/sso/auth/mocks/token-hasher"
	mocks_tokener "github.com/eragon-mdi/sso/internal/service/sso/auth/mocks/tokener"
)

func baseCfg() *configs.BussinesLogic {
	return &configs.BussinesLogic{
		TokenTTL: time.Hour,
	}
}

func TestRegister_AllCases(t *testing.T) {
	ctx := context.Background()
	inUser := domain.User{
		Email:    "u@e.x",
		Password: "pass",
	}

	t.Run("success", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		// NewUser возвращает тот же объект (ID устанавливает Register)
		repo.On("NewUser", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
			return u.Email == inUser.Email && u.Password != inUser.Password // пароль должен быть уже захеширован
		})).Return(func(_ context.Context, u domain.User) domain.User { return u }, nil)

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Gen", mock.Anything).Return([]byte("hashed-pass"), nil)

		s := New(repo, hasher, nil, nil, baseCfg())

		got, err := s.Register(ctx, inUser)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.Email != inUser.Email {
			t.Fatalf("email mismatch: got %q want %q", got.Email, inUser.Email)
		}
		hasher.AssertCalled(t, "Gen", mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("hash fail", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Gen", mock.Anything).Return([]byte(nil), errors.New("hash fail"))

		s := New(repo, hasher, nil, nil, baseCfg())
		_, err := s.Register(ctx, inUser)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		hasher.AssertExpectations(t)
	})

	t.Run("duplicate email -> wrap domain.ErrDuplicate", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("NewUser", mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrDuplicate)

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Gen", mock.Anything).Return([]byte("ok"), nil)

		s := New(repo, hasher, nil, nil, baseCfg())
		_, err := s.Register(ctx, inUser)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !errors.Is(err, domain.ErrDuplicate) {
			t.Fatalf("expected wrapped domain.ErrDuplicate; got: %v", err)
		}
	})

	t.Run("repo other error", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("NewUser", mock.Anything, mock.Anything).Return(domain.User{}, errors.New("db boom"))

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Gen", mock.Anything).Return([]byte("ok"), nil)

		s := New(repo, hasher, nil, nil, baseCfg())
		_, err := s.Register(ctx, inUser)
		if err == nil {
			t.Fatal("expected repo error")
		}
	})
}

func TestLogin_AllCases(t *testing.T) {
	ctx := context.Background()

	stored := domain.User{
		ID:       "uid",
		Email:    "e@x.y",
		Password: "stored-hash",
	}

	// dctx: int32 fields
	dctx := domain.DeviceCtx{AppId: int32(10), DeviceID: int32(20)}

	t.Run("success", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("GetUserInfoByEmail", mock.Anything, stored.Email).Return(stored, nil)
		repo.On("SaveRefreshToken", mock.Anything, mock.Anything).Return(nil)

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Compare", mock.Anything, mock.Anything).Return(true, nil)

		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte("acc"), []byte("ref"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("href"), nil)

		s := New(repo, hasher, tokener, tokenHasher, baseCfg())

		got, err := s.Login(ctx, domain.User{Email: stored.Email, Password: "plain"}, dctx)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.Access == "" || got.Refresh == "" {
			t.Fatalf("expected tokens not empty")
		}

		repo.AssertExpectations(t)
		hasher.AssertExpectations(t)
		tokener.AssertExpectations(t)
		tokenHasher.AssertExpectations(t)
	})

	t.Run("get user error", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("GetUserInfoByEmail", mock.Anything, mock.Anything).Return(domain.User{}, errors.New("no user"))

		hasher := &mocks_hasher.PasswordHasher{}
		s := New(repo, hasher, nil, nil, baseCfg())

		_, err := s.Login(ctx, domain.User{Email: "x"}, dctx)
		if err == nil {
			t.Fatal("expected error when GetUserInfoByEmail fails")
		}
	})

	t.Run("compare error or wrong pass", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("GetUserInfoByEmail", mock.Anything, mock.Anything).Return(stored, nil)

		hasher := &mocks_hasher.PasswordHasher{}
		// simulate wrong password
		hasher.On("Compare", mock.Anything, mock.Anything).Return(false, nil)

		s := New(repo, hasher, nil, nil, baseCfg())
		_, err := s.Login(ctx, domain.User{Email: stored.Email, Password: "bad"}, dctx)
		if err == nil {
			t.Fatal("expected error on wrong password")
		}
	})

	t.Run("tokener generation error", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("GetUserInfoByEmail", mock.Anything, mock.Anything).Return(stored, nil)
		repo.On("SaveRefreshToken", mock.Anything, mock.Anything).Return(nil) // won't be called but safe

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Compare", mock.Anything, mock.Anything).Return(true, nil)

		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte(nil), []byte(nil), errors.New("jwt fail"))

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		s := New(repo, hasher, tokener, tokenHasher, baseCfg())
		_, err := s.Login(ctx, domain.User{Email: stored.Email, Password: "ok"}, dctx)
		if err == nil {
			t.Fatal("expected error when tokener.GenPair fails")
		}
	})

	t.Run("save refresh token error", func(t *testing.T) {
		repo := &mocks_repo.Repository{}
		repo.On("GetUserInfoByEmail", mock.Anything, mock.Anything).Return(stored, nil)
		repo.On("SaveRefreshToken", mock.Anything, mock.Anything).Return(errors.New("save fail"))

		hasher := &mocks_hasher.PasswordHasher{}
		hasher.On("Compare", mock.Anything, mock.Anything).Return(true, nil)

		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte("a"), []byte("r"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		s := New(repo, hasher, tokener, tokenHasher, baseCfg())
		_, err := s.Login(ctx, domain.User{Email: stored.Email, Password: "ok"}, dctx)
		if err == nil {
			t.Fatal("expected error when SaveRefreshToken fails")
		}
	})
}

func TestVerificationTokenAndGenFlow(t *testing.T) {
	// проверим verificationToken и genTokensFlow частично (различные ошибки и успех)
	//ctx := context.Background()
	userDctx := domain.NewDeviceCtx(int32(1), int32(2))
	// validMeta := domain.NewMeta(time.Hour, "user-1", userDctx.AppId, userDctx.DeviceID)

	t.Run("verificationToken invalid token", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(domain.Meta{}, errors.New("bad"))

		s := New(nil, nil, tokener, nil, baseCfg())
		_, err := s.verificationToken("bad", userDctx)
		if err == nil {
			t.Fatal("expected error for invalid token")
		}
	})

	t.Run("verificationToken ctx mismatch", func(t *testing.T) {
		// tokener returns meta with different ctx
		meta := domain.NewMeta(time.Hour, "u", int32(9), int32(9))
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(meta, nil)

		s := New(nil, nil, tokener, nil, baseCfg())
		_, err := s.verificationToken("tok", userDctx)
		if err == nil {
			t.Fatal("expected ctx mismatch error")
		}
	})

	t.Run("genTokensFlow: tokener.GenPair fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte(nil), []byte(nil), errors.New("gen fail"))

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		_, _, err := s.genTokensFlow("uid", userDctx)
		if err == nil {
			t.Fatal("expected tokener gen error")
		}
	})

	t.Run("genTokensFlow: tokenHasher fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte("a"), []byte("r"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte(nil), errors.New("sum fail"))

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		_, _, err := s.genTokensFlow("uid", userDctx)
		if err == nil {
			t.Fatal("expected tokenHasher sum error")
		}
	})

	t.Run("genTokensFlow success", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("GenPair", mock.Anything).Return([]byte("acc"), []byte("ref"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("hashref"), nil)

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		tok, rt, err := s.genTokensFlow("uid", userDctx)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if tok == nil || rt == nil {
			t.Fatal("expected non-nil outputs")
		}
		if tok.Access == "" || tok.Refresh == "" {
			t.Fatal("tokens empty")
		}
		if rt.Token == "" {
			t.Fatal("refresh token hash empty")
		}
	})
}

func TestRefreshAndLogout_AllCases(t *testing.T) {
	ctx := context.Background()
	userDctx := domain.NewDeviceCtx(int32(5), int32(7))
	validMeta := domain.NewMeta(time.Hour, "u1", userDctx.AppId, userDctx.DeviceID)

	t.Run("Refresh verify token fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(domain.Meta{}, errors.New("invalid"))
		s := New(nil, nil, tokener, nil, baseCfg())
		_, err := s.Refresh(ctx, "old", userDctx)
		if err == nil {
			t.Fatal("expected verify error")
		}
	})

	t.Run("Refresh gen tokens fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)
		tokener.On("GenPair", mock.Anything).Return([]byte(nil), []byte(nil), errors.New("gen fail"))

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		_, err := s.Refresh(ctx, "old", userDctx)
		if err == nil {
			t.Fatal("expected gen tokens error")
		}
	})

	t.Run("Refresh tokenHasher sum fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)
		tokener.On("GenPair", mock.Anything).Return([]byte("a"), []byte("r"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte(nil), errors.New("sum fail"))

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		_, err := s.Refresh(ctx, "old", userDctx)
		if err == nil {
			t.Fatal("expected sum error")
		}
	})

	t.Run("Refresh rotate token fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)
		tokener.On("GenPair", mock.Anything).Return([]byte("a"), []byte("r"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		repo := &mocks_repo.Repository{}
		repo.On("RotateToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("rotate fail"))

		s := New(repo, nil, tokener, tokenHasher, baseCfg())
		_, err := s.Refresh(ctx, "old", userDctx)
		if err == nil {
			t.Fatal("expected rotate error")
		}
	})

	t.Run("Refresh tokenHasher.Sum fail", func(t *testing.T) {
		ctx := context.Background()
		userDctx := domain.NewDeviceCtx(int32(5), int32(7))
		validMeta := domain.NewMeta(time.Hour, "u1", userDctx.AppId, userDctx.DeviceID)

		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)
		tokener.On("GenPair", mock.Anything).Return([]byte("access"), []byte("refresh"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte(nil), errors.New("sum fail"))

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		_, err := s.Refresh(ctx, "old-refresh", userDctx)
		if err == nil {
			t.Fatal("expected error when tokenHasher.Sum fails")
		}
	})

	t.Run("Refresh success", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)
		tokener.On("GenPair", mock.Anything).Return([]byte("a"), []byte("r"), nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		repo := &mocks_repo.Repository{}
		repo.On("RotateToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		s := New(repo, nil, tokener, tokenHasher, baseCfg())
		got, err := s.Refresh(ctx, "old", userDctx)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.Access == "" || got.Refresh == "" {
			t.Fatal("tokens empty")
		}
	})

	t.Run("Logout verify fail", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(domain.Meta{}, errors.New("invalid"))
		s := New(nil, nil, tokener, nil, baseCfg())
		err := s.Logout(ctx, "r", userDctx)
		if err == nil {
			t.Fatal("expected verify error on logout")
		}
	})

	t.Run("Logout sum error", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte(nil), errors.New("sum fail"))

		s := New(nil, nil, tokener, tokenHasher, baseCfg())
		err := s.Logout(ctx, "r", userDctx)
		if err == nil {
			t.Fatal("expected hashing error")
		}
	})

	t.Run("Logout revoke returns ErrNotFound => ignored", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		repo := &mocks_repo.Repository{}
		repo.On("RevokeTokenByHash", mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		s := New(repo, nil, tokener, tokenHasher, baseCfg())
		if err := s.Logout(ctx, "r", userDctx); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})

	t.Run("Logout revoke returns error -> bubbled", func(t *testing.T) {
		tokener := &mocks_tokener.Tokener{}
		tokener.On("VerifyRefresh", mock.Anything).Return(validMeta, nil)

		tokenHasher := &mocks_tokenhasher.TokenHasher{}
		tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

		repo := &mocks_repo.Repository{}
		repo.On("RevokeTokenByHash", mock.Anything, mock.Anything).Return(errors.New("boom"))

		s := New(repo, nil, tokener, tokenHasher, baseCfg())
		if err := s.Logout(ctx, "r", userDctx); err == nil {
			t.Fatal("expected revoke error propagated")
		}
	})
}

// sanity check internal functions behaviour (types/values)
func Test_internalSanity(t *testing.T) {
	// simple sanity: genTokensFlow + verificationToken round-ish checks
	userDctx := domain.NewDeviceCtx(int32(2), int32(3))
	tokener := &mocks_tokener.Tokener{}
	tokener.On("GenPair", mock.Anything).Return([]byte("acc"), []byte("ref"), nil)
	tokener.On("VerifyRefresh", mock.Anything).Return(domain.NewMeta(time.Minute, "u1", userDctx.AppId, userDctx.DeviceID), nil)

	tokenHasher := &mocks_tokenhasher.TokenHasher{}
	tokenHasher.On("Sum", mock.Anything).Return([]byte("h"), nil)

	s := New(nil, nil, tokener, tokenHasher, baseCfg())

	tok, rt, err := s.genTokensFlow("u1", userDctx)
	if err != nil {
		t.Fatalf("genTokensFlow err: %v", err)
	}
	if reflect.DeepEqual(tok, (*domain.Token)(nil)) || reflect.DeepEqual(rt, (*domain.RefreshToken)(nil)) {
		t.Fatalf("unexpected nil results")
	}

	// verificationToken will use tokener.VerifyRefresh mocked above
	_, err = s.verificationToken("some", userDctx)
	if err != nil {
		t.Fatalf("verificationToken unexpected err: %v", err)
	}
}
