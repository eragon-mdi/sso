package grpctransportauth

import (
	"context"
	"errors"
	"testing"

	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/domain"
	mocks "github.com/eragon-mdi/sso/internal/transport/http2/grpc/sso/auth/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthTransport_Register(t *testing.T) {
	ctx := context.Background()
	user := &sso.User{Email: "a@b.c", Password: "123456"}

	t.Run("invalid request", func(t *testing.T) {
		srv := New(&mocks.AuthService{}, zap.NewNop().Sugar())
		_, err := srv.Register(ctx, &sso.RegisterRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("service duplicate error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Register", mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrDuplicate)

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Register(ctx, &sso.RegisterRequest{User: user})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.AlreadyExists, st.Code())

		s.AssertExpectations(t)
	})

	t.Run("service internal error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Register", mock.Anything, mock.Anything).Return(domain.User{}, errors.New("boom"))

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Register(ctx, &sso.RegisterRequest{User: user})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Register", mock.Anything, mock.Anything).Return(domain.User{ID: "123"}, nil)

		srv := New(s, zap.NewNop().Sugar())
		resp, err := srv.Register(ctx, &sso.RegisterRequest{
			User: &sso.User{
				Email:    "test@example.com",
				Password: "123456",
			},
		})

		require.NoError(t, err)
		require.Equal(t, "123", resp.UserId)
		s.AssertExpectations(t)
	})
}

func TestAuthTransport_Login(t *testing.T) {
	ctx := context.Background()
	user := &sso.User{Email: "a@b.c", Password: "123456"}
	device := &sso.DeviceContext{AppId: 1, DeviceId: 2}
	token := domain.Token{Access: "acc", Refresh: "ref"}

	t.Run("invalid request", func(t *testing.T) {
		srv := New(&mocks.AuthService{}, zap.NewNop().Sugar())
		_, err := srv.Login(ctx, &sso.LoginRequest{})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("service validation error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(domain.Token{}, domain.ErrValidation)

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Login(ctx, &sso.LoginRequest{User: user, Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("service internal error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(domain.Token{}, errors.New("boom"))

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Login(ctx, &sso.LoginRequest{User: user, Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(token, nil)

		srv := New(s, zap.NewNop().Sugar())
		resp, err := srv.Login(ctx, &sso.LoginRequest{User: user, Ctx: device})
		require.NoError(t, err)
		require.Equal(t, "acc", resp.Tokens.Access)
		require.Equal(t, "ref", resp.Tokens.Refresh)
	})
}

func TestAuthTransport_Refresh(t *testing.T) {
	ctx := context.Background()
	device := &sso.DeviceContext{AppId: 1, DeviceId: 2}
	token := domain.Token{Access: "acc", Refresh: "ref"}

	t.Run("invalid request", func(t *testing.T) {
		srv := New(&mocks.AuthService{}, zap.NewNop().Sugar())
		_, err := srv.Refresh(ctx, &sso.RefreshRequest{})
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("service validation error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Refresh", mock.Anything, "r", mock.Anything).Return(domain.Token{}, domain.ErrValidation)

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Refresh(ctx, &sso.RefreshRequest{Refresh: "r", Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("service internal error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Refresh", mock.Anything, "r", mock.Anything).Return(domain.Token{}, errors.New("boom"))

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Refresh(ctx, &sso.RefreshRequest{Refresh: "r", Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Refresh", mock.Anything, "r", mock.Anything).Return(token, nil)

		srv := New(s, zap.NewNop().Sugar())
		resp, err := srv.Refresh(ctx, &sso.RefreshRequest{Refresh: "r", Ctx: device})
		require.NoError(t, err)
		require.Equal(t, "acc", resp.Tokens.Access)
		require.Equal(t, "ref", resp.Tokens.Refresh)
	})
}

func TestAuthTransport_Logout(t *testing.T) {
	ctx := context.Background()
	device := &sso.DeviceContext{AppId: 1, DeviceId: 2}

	t.Run("invalid request", func(t *testing.T) {
		srv := New(&mocks.AuthService{}, zap.NewNop().Sugar())
		_, err := srv.Logout(ctx, &sso.LogoutRequest{})
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("service validation error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Logout", mock.Anything, "r", mock.Anything).Return(domain.ErrValidation)

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Logout(ctx, &sso.LogoutRequest{Refresh: "r", Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("service internal error", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Logout", mock.Anything, "r", mock.Anything).Return(errors.New("boom"))

		srv := New(s, zap.NewNop().Sugar())
		_, err := srv.Logout(ctx, &sso.LogoutRequest{Refresh: "r", Ctx: device})
		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		s := &mocks.AuthService{}
		s.On("Logout", mock.Anything, "r", mock.Anything).Return(nil)

		srv := New(s, zap.NewNop().Sugar())
		resp, err := srv.Logout(ctx, &sso.LogoutRequest{Refresh: "r", Ctx: device})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestValidate_RequestStructs(t *testing.T) {
	// 1. RegisterRequest без User
	err := validate(&sso.RegisterRequest{})
	require.Error(t, err)
	require.Equal(t, "user is required", err.Error())

	// 2. LoginRequest без User
	err = validate(&sso.LoginRequest{
		Ctx: &sso.DeviceContext{AppId: 1, DeviceId: 1},
	})
	require.Error(t, err)
	require.Equal(t, "user is required", err.Error())

	// 3. LoginRequest без Ctx
	err = validate(&sso.LoginRequest{
		User: &sso.User{Email: "a@b.c", Password: "123456"},
	})
	require.Error(t, err)
	require.Equal(t, "device context is required", err.Error())

	// 4. LogoutRequest без Ctx
	err = validate(&sso.LogoutRequest{
		Refresh: "token",
	})
	require.Error(t, err)
	require.Equal(t, "device context is required", err.Error())

	// 5. RefreshRequest без Ctx
	err = validate(&sso.RefreshRequest{
		Refresh: "token",
	})
	require.Error(t, err)
	require.Equal(t, "device context is required", err.Error())

	// 6. Неизвестный тип
	err = validate("string")
	require.Error(t, err)
	require.Equal(t, "bad request type", err.Error())
}
