package grpctransportauth

import (
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/common/api"
	"go.uber.org/zap"
)

type authTransport struct {
	s AuthService
	l *zap.SugaredLogger
	sso.UnimplementedAuthServer
}

func New(s AuthService, l *zap.SugaredLogger) api.AuthTransport {
	return &authTransport{
		s: s,
		l: l,
	}
}
