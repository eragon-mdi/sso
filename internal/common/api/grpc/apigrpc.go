package apigrpc

import "github.com/eragon-mdi/protos/gen/go/sso/v1"

type GrpcApi interface {
	sso.AuthServer
}

func New()  {
	
}
