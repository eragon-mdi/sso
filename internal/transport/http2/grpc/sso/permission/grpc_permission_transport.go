package grpctransportpermission

import (
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/common/api"
	"go.uber.org/zap"
)

type permissionTransport struct {
	s PermissionService
	l *zap.SugaredLogger
	sso.UnimplementedPermissionServer
}

func New(s PermissionService, l *zap.SugaredLogger) api.PermissionTransport {
	return &permissionTransport{
		s: s,
		l: l,
	}
}
