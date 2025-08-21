package api

import (
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/common/server"
	"google.golang.org/grpc/reflection"
)

type Transport interface {
	AuthTransport
	PermissionTransport
}

type AuthTransport interface {
	sso.AuthServer
}

type PermissionTransport interface {
	sso.PermissionServer
}

func RegisterRoutes(s server.Server, t Transport) {
	// grpc
	sso.RegisterAuthServer(s.GRPC(), t)
	sso.RegisterPermissionServer(s.GRPC(), t)

	reflection.Register(s.GRPC())
}
