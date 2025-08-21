package grpctransportpermission

import (
	"context"

	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=PermissionService --with-expecter --output=./mocks --exported
type PermissionService interface {
	IsAdmin(context.Context, domain.User) (bool, error)
}

const (
	ErrFailedValidateReq  = "failed to validate request"
	FailedCheckIsAdminReq = "failed to check is user admin"
)

func (t permissionTransport) IsAdmin(ctx context.Context, req *sso.IsAdminRequest) (*sso.IsAdminResponse, error) {
	if err := validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrFailedValidateReq)
	}

	isAdmin, err := t.s.IsAdmin(ctx, userFromIsAdminReq(req))
	if err != nil {
		return nil, status.Error(codes.Internal, FailedCheckIsAdminReq)
	}

	return &sso.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
