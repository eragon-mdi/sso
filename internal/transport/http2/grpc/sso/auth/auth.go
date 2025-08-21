package grpctransportauth

import (
	"context"
	"errors"

	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockery --name=AuthService --with-expecter --output=./mocks --exported
type AuthService interface {
	Register(context.Context, domain.User) (domain.User, error)
	Login(context.Context, domain.User, domain.DeviceCtx) (domain.Token, error)
	Refresh(context.Context, string, domain.DeviceCtx) (domain.Token, error)
	Logout(context.Context, string, domain.DeviceCtx) error
}

const (
	ErrFailedValidateReq = "failed to validate request"
	ErrFailedRegisterReq = "failed to register user"
	ErrFailedLoginReq    = "failed to login user"
	ErrFailedRefreshReq  = "failed to refrsh user token"
	ErrFailedLogoutReq   = "failed to logout user"
)

func (t authTransport) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	if err := validate(req); err != nil {
		t.l.Errorw(ErrFailedValidateReq, err)
		return nil, status.Error(codes.InvalidArgument, ErrFailedValidateReq)
	}

	user, err := t.s.Register(ctx, userFromRegisterReq(req))
	if err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			t.l.Errorw(ErrFailedRegisterReq, err)
			return nil, status.Error(codes.AlreadyExists, ErrFailedRegisterReq)
		}
		t.l.Errorw(ErrFailedRegisterReq, err)
		return nil, status.Error(codes.Internal, ErrFailedRegisterReq)
	}

	return &sso.RegisterResponse{
		UserId: user.ID,
	}, nil
}

func (t authTransport) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	if err := validate(req); err != nil {
		t.l.Errorw(ErrFailedValidateReq, err)
		return nil, status.Error(codes.InvalidArgument, ErrFailedValidateReq)
	}

	token, err := t.s.Login(ctx, userFromLoginReq(req), deviceCtxFromReq(req.Ctx))
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			t.l.Errorw(ErrFailedLoginReq, err)
			return nil, status.Error(codes.Unauthenticated, ErrFailedLoginReq)
		}
		t.l.Errorw(ErrFailedLoginReq, err)
		return nil, status.Error(codes.Internal, ErrFailedLoginReq)
	}

	return &sso.LoginResponse{
		Tokens: tokenResponse(token),
	}, nil
}

func (t authTransport) Refresh(ctx context.Context, req *sso.RefreshRequest) (*sso.RefreshResponse, error) {
	if err := validate(req); err != nil {
		t.l.Errorw(ErrFailedValidateReq, err)
		return nil, status.Error(codes.InvalidArgument, ErrFailedValidateReq)
	}

	token, err := t.s.Refresh(ctx, req.Refresh, deviceCtxFromReq(req.Ctx))
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			t.l.Errorw(ErrFailedRefreshReq, err)
			return nil, status.Error(codes.Unauthenticated, ErrFailedRefreshReq)
		}
		t.l.Errorw(ErrFailedRefreshReq, err)
		return nil, status.Error(codes.Internal, ErrFailedRefreshReq)
	}

	return &sso.RefreshResponse{
		Tokens: tokenResponse(token),
	}, nil
}

func (t authTransport) Logout(ctx context.Context, req *sso.LogoutRequest) (*emptypb.Empty, error) {
	if err := validate(req); err != nil {
		t.l.Errorw(ErrFailedValidateReq, err)
		return nil, status.Error(codes.InvalidArgument, ErrFailedValidateReq)
	}

	if err := t.s.Logout(ctx, req.Refresh, deviceCtxFromReq(req.Ctx)); err != nil {
		if errors.Is(err, domain.ErrValidation) {
			t.l.Errorw(ErrFailedLogoutReq, err)
			return nil, status.Error(codes.Unauthenticated, ErrFailedLogoutReq)
		}
		t.l.Errorw(ErrFailedLogoutReq, err)
		return nil, status.Error(codes.Internal, ErrFailedLogoutReq)
	}

	return &emptypb.Empty{}, nil
}

func tokenResponse(token domain.Token) *sso.TokenPair {
	return &sso.TokenPair{
		Refresh: token.Refresh,
		Access:  token.Access,
	}
}
