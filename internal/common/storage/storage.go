package storage

import (
	"context"
	"time"

	pgdriver "github.com/eragon-mdi/go-playground/storage/drivers/postgres"
	redisstore "github.com/eragon-mdi/go-playground/storage/nosql/redis"
	sqlstore "github.com/eragon-mdi/go-playground/storage/sql"
	"github.com/eragon-mdi/sso/internal/common/configs"
	"github.com/go-faster/errors"
)

const (
	ErrConnectDB       = "Failed to connect to db"
	ErrDisconnectSqlDB = "Failed to disconnect sql-db"
	ConnTimeoutDefault = time.Minute
)

type Storage interface {
	SQL() sqlstore.Storage
	Redis() redisstore.Storage
	GracefulShutdown() error
}

type storage struct {
	sqlStore   sqlstore.Storage
	redisStore redisstore.Storage
}

func Conn(ctx context.Context, cfg *configs.Storages, timeout time.Duration) (Storage, error) {
	sql, err := sqlstore.Conn(ctx, cfg.Postgres, pgdriver.Postgres{}, timeout)
	if err != nil {
		return nil, errors.Wrap(err, ErrConnectDB)
	}

	redis, err := redisstore.Conn(ctx, cfg.Redis, timeout)
	if err != nil {
		return nil, errors.Wrap(err, ErrConnectDB)
	}

	return &storage{
		sqlStore:   sql,
		redisStore: redis,
	}, nil
}

func (s storage) SQL() sqlstore.Storage {
	return s.sqlStore
}

func (s storage) Redis() redisstore.Storage {
	return s.redisStore
}

func (s storage) GracefulShutdown() error {
	if err := s.sqlStore.Close(); err != nil {
		return errors.Wrap(err, ErrDisconnectSqlDB)
	}

	return nil
}
