package srvgrpc

import (
	"fmt"
	"net"

	"github.com/eragon-mdi/sso/internal/common/configs"
	"github.com/go-faster/errors"
	"google.golang.org/grpc"
)

type GrpcSrv struct {
	*grpc.Server

	port string
}

func New(cfg configs.Server) *GrpcSrv {
	return &GrpcSrv{
		Server: grpc.NewServer(),
		port:   cfg.PortF,
	}
}

// Try open listener port from cfg && start grpc-srv
// Use in gorutine!
func (s *GrpcSrv) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return errors.Wrap(err, "failed listen port:")
	}

	if err := s.Server.Serve(lis); err != nil {
		return errors.Wrap(err, "failed start grpcSrv:")
	}

	return nil
}
