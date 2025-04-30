package redis

import (
	"github.com/go-redis/redis"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository() (*RedisRepository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	rr := &RedisRepository{
		client: rdb,
	}

	err := rr.Ping()
	if err != nil {
		return nil, err
	}
	return rr, nil
}

func (r *RedisRepository) Ping() error {
	return r.client.Ping().Err()
}
