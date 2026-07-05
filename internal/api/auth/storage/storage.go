package storage

import (
	"context"
	"time"

	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(
		ctx context.Context,
		key string,
		value interface{},
		expiration time.Duration,
	) *redis.StatusCmd

	Del(
		ctx context.Context,
		keys ...string,
	) *redis.IntCmd

	GetDel(
		ctx context.Context,
		key string,
	) *redis.StringCmd
}

type Repository struct {
	rdb Cache
}

func New(
	rdb Cache,
) *Repository {

	return &Repository{
		rdb: rdb,
	}
}

func (r *Repository) Save(
	ctx context.Context,
	tokenID string,
	ttl time.Duration,
) error {

	return r.rdb.Set(
		ctx,
		"refresh:"+tokenID,
		1,
		ttl,
	).Err()
}

func (r *Repository) Delete(
	ctx context.Context,
	tokenID string,
) error {

	return r.rdb.Del(
		ctx,
		"refresh:"+tokenID,
	).Err()
}

func (r *Repository) Consume(
	ctx context.Context,
	tokenID string,
) error {

	res, err := r.rdb.GetDel(
		ctx,
		"refresh:"+tokenID,
	).Result()

	if err == redis.Nil {
		return domain.ErrTokenRevoked
	}

	if err != nil {
		return err
	}

	if res == "" {
		return domain.ErrTokenRevoked
	}

	return nil
}
