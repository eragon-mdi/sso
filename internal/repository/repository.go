package repository

import (
	"github.com/eragon-mdi/sso/internal/common/storage"
	redisrepo "github.com/eragon-mdi/sso/internal/repository/nosql/redis"
	sqlrepo "github.com/eragon-mdi/sso/internal/repository/sql"
	"github.com/eragon-mdi/sso/internal/service"
)

func New(s storage.Storage) service.Repository {
	return &repository{
		SqlRepo:   sqlrepo.New(s.SQL()),
		RedisRepo: redisrepo.New(s.Redis()),
	}
}

type repository struct {
	sqlrepo.SqlRepo
	redisrepo.RedisRepo
}
