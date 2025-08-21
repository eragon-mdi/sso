package transport

import (
	"github.com/eragon-mdi/sso/internal/common/api"
	grpctransportauth "github.com/eragon-mdi/sso/internal/transport/http2/grpc/sso/auth"
	grpctransportpermission "github.com/eragon-mdi/sso/internal/transport/http2/grpc/sso/permission"
	"go.uber.org/zap"
)

type Service interface {
	grpctransportauth.AuthService
	grpctransportpermission.PermissionService
}

type transport struct {
	api.AuthTransport
	api.PermissionTransport
}

func New(s Service, l *zap.SugaredLogger) api.Transport {
	return &transport{
		AuthTransport:       grpctransportauth.New(s, l),
		PermissionTransport: grpctransportpermission.New(s, l),
	}
}
