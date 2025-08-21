package main

import (
	"log"

	"github.com/eragon-mdi/sso/internal/common/api"
	"github.com/eragon-mdi/sso/internal/common/configs"
	"github.com/eragon-mdi/sso/internal/common/logger"
	"github.com/eragon-mdi/sso/internal/common/server"
	"github.com/eragon-mdi/sso/internal/common/storage"
	"github.com/eragon-mdi/sso/internal/repository"
	"github.com/eragon-mdi/sso/internal/service"
	"github.com/eragon-mdi/sso/internal/transport"

	rootctx "github.com/eragon-mdi/go-playground/server/root-ctx"
)

func main() {
	cfg := configs.MustLoad()

	l, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("failed to set logger: %v", err)
	}

	ctx, cancelAppCtx := rootctx.NotifyBackgroundCtxToShutdownSignal()
	defer cancelAppCtx()

	store, err := storage.Conn(ctx, &cfg.Storages, storage.ConnTimeoutDefault)
	if err != nil {
		l.Error(err)
		return
	}

	r := repository.New(store)
	s, err := service.New(r, &cfg.BussinesLogic)
	if err != nil {
		l.Error(err)
		cancelAppCtx()
		return
	}
	t := transport.New(s, l)

	srv := server.New(&cfg.Servers)
	api.RegisterRoutes(srv, t)
	go func() {
		if err := srv.StartAll(); err != nil {
			l.Errorf("failed to start servers: %v", err)
			cancelAppCtx()
		}
	}()

	<-ctx.Done()

	if err := srv.GracefulShutdown(); err != nil {
		l.Errorw("error during server shutdown", "cause", err)
	}
	if err := store.GracefulShutdown(); err != nil {
		l.Errorw("error disconnect store", "cause", err)
	}
}
