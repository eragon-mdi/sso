package server

import (
	"github.com/eragon-mdi/sso/internal/common/configs"
	srvgrpc "github.com/eragon-mdi/sso/internal/common/server/grpc"
	"golang.org/x/sync/errgroup"
)

type Server interface {
	StartAll() error
	GracefulShutdown() error

	GRPC() *srvgrpc.GrpcSrv
	//todo other servers, example REST
}

type server struct {
	grpc *srvgrpc.GrpcSrv
}

func New(cfg *configs.Servers) Server {
	return &server{
		grpc: srvgrpc.New(cfg.GRPC),
	}
}

func (s *server) StartAll() error {
	eg := errgroup.Group{}
	defer func() {
		_ = eg.Wait()
	}()

	eg.Go(func() error {
		return s.GRPC().Serve()
	})

	//eg.Go(func() error {
	//	return http.ListenAndServe("0.0.0.0:7070", nil)
	//})

	return eg.Wait()
}

func (s *server) GRPC() *srvgrpc.GrpcSrv {
	return s.grpc
}

// waiting all
func (s *server) GracefulShutdown() error {
	s.grpc.GracefulStop()

	// if err := s.rest.Shutdown(context.Background()); err != nil {
	// return err
	// }

	return nil
}
