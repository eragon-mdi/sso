package redisrepo

import (
	redisstore "github.com/eragon-mdi/go-playground/storage/nosql/redis"
	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
)

type RedisRepo interface {
	authservice.TokenRepository
}

type redisRepo struct {
	s redisstore.Storage
}

func New(s redisstore.Storage) RedisRepo {
	return &redisRepo{
		s: s,
	}
}
