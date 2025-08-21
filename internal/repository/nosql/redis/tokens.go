package redisrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/go-faster/errors"
)

func (r *redisRepo) SaveRefreshToken(ctx context.Context, rt domain.RefreshToken) error {
	val, err := json.Marshal(newTokenMeta(rt.Meta))
	if err != nil {
		return errors.Wrap(err, "failed prepare data: marshal")
	}

	ttl := time.Until(rt.Meta.Exp)
	// минимальный TTL, чтобы ключ не жил вечно, если exp «в прошлом»
	if ttl <= 0 {
		ttl = time.Second
	}

	ok, err := r.s.SetNX(ctx, key(rt.Token), val, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrDuplicate
	}

	return nil
}

func (r *redisRepo) RotateToken(ctx context.Context, oldHash string, rt domain.RefreshToken) error {
	val, err := json.Marshal(newTokenMeta(rt.Meta))
	if err != nil {
		return errors.Wrap(err, "redis: marshal new token meta")
	}

	ttl := time.Until(rt.Meta.Exp)
	if ttl <= 0 {
		ttl = time.Second
	}
	ttlMs := int64(ttl / time.Millisecond)

	res, err := r.s.Eval(ctx, rotateTokenLua, []string{key(oldHash), key(rt.Token)}, val, ttlMs).Int()
	if err != nil {
		return errors.Wrap(err, "redis: eval rotateTokenLua")
	}
	switch res {
	case 1:
		return nil
	case 0:
		return domain.ErrNotFound
	case -1:
		return domain.ErrDuplicate
	default:
		return errors.New("redis: unexpected result from rotate script")
	}
}

func (r *redisRepo) RevokeTokenByHash(ctx context.Context, hash string) error {
	n, err := r.s.Del(ctx, key(hash)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func key(hash string) string { return "rt:" + hash }
