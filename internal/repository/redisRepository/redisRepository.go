package redisRepository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/vakhrushevk/cloudru/internal/repository"
	"github.com/vakhrushevk/cloudru/internal/repository/model"
)

var (
	ErrRedisClientNil = errors.New("redis client is nil")
	ErrBucketNotFound = errors.New("bucket not found")
)

type BucketRepository struct {
	client *redis.Client
}

func NewRedisRepository(redis *redis.Client) (repository.BucketRepository, error) {
	if redis == nil {
		return nil, ErrRedisClientNil
	}
	return &BucketRepository{
		client: redis,
	}, nil
}

func bucketKey(key string) string {
	return fmt.Sprintf("ratelimit:bucket:%s", key)
}

func (r *BucketRepository) CreateBucket(ctx context.Context, key string, capacity int, refilRate int, tokens int) error {
	now := time.Now()

	bucket := map[string]interface{}{
		"tokens":      tokens,
		"capacity":    capacity,
		"refil_rate":  refilRate,
		"last_refill": now.Unix(),
	}

	pipe := r.client.Pipeline()
	pipe.HMSet(bucketKey(key), bucket)

	_, err := pipe.Exec()
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (r *BucketRepository) RefillAllBuckets(ctx context.Context) error {
	var cursor uint64
	now := time.Now().Unix()

	script := `
        local current_tokens = tonumber(redis.call('HGET', KEYS[1], 'tokens'))
        local capacity = tonumber(redis.call('HGET', KEYS[1], 'capacity'))
        local refil_rate = tonumber(redis.call('HGET', KEYS[1], 'refil_rate'))
        local last_refill = tonumber(redis.call('HGET', KEYS[1], 'last_refill'))
        
        if not (current_tokens and capacity and refil_rate and last_refill) then
            return 0
        end
        
        local elapsed = tonumber(ARGV[1]) - last_refill
        local added_tokens = math.floor(elapsed * refil_rate)
        local new_tokens = math.min(capacity, current_tokens + added_tokens)
        
        if new_tokens > current_tokens then
            redis.call('HMSET', KEYS[1],
                'tokens', new_tokens,
                'last_refill', ARGV[1]
            )
            return 1
        end
        return 0
    `

	for {
		keys, newCursor, err := r.client.Scan(cursor, "ratelimit:bucket:*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan redis: %w", err)
		}
		if len(keys) > 0 {
			pipe := r.client.Pipeline()

			for _, key := range keys {
				pipe.Eval(script, []string{key}, now)
			}

			cmders, err := pipe.Exec()
			if err != nil {
				return fmt.Errorf("failed to execute pipeline: %w", err)
			}

			for i, cmder := range cmders {
				if err := cmder.Err(); err != nil {
					slog.Error("Failed to refill bucket",
						"key", keys[i],
						"error", err)
				}
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}
	}

	return nil
}

func (r *BucketRepository) Decrease(ctx context.Context, key string) (bool, error) {
	script := `
        local data = redis.call('HMGET', KEYS[1], 'tokens', 'last_refill', 'capacity', 'refil_rate')
        if not data[1] then
            return {err = "NOT_FOUND"}
        end
        
        local current_tokens = tonumber(data[1])
        local last_refill = tonumber(data[2])
        local capacity = tonumber(data[3])
        local refil_rate = tonumber(data[4])
        
        local now = tonumber(ARGV[1])
        local elapsed = now - last_refill
        
        -- Вычисляем текущее количество токенов с учетом пополнения
        local added_tokens = math.floor(elapsed * refil_rate)
        local available_tokens = math.min(capacity, current_tokens + added_tokens)
        
        -- Проверяем, есть ли хотя бы 1 токен
        if available_tokens >= 1 then
            redis.call('HMSET', KEYS[1], 
                'tokens', available_tokens - 1,
                'last_refill', now
            )
            return 1
        end
        
        -- Обновляем время в любом случае
        redis.call('HSET', KEYS[1], 'last_refill', now)
        return 0
    `

	result, err := r.client.Eval(script, []string{bucketKey(key)}, time.Now().Unix()).Result()
	if err != nil {
		return false, fmt.Errorf("failed to decrease tokens: %w", err)
	}

	// Изменена обработка результата
	switch v := result.(type) {
	case int64:
		return v == 1, nil
	case []interface{}:
		if len(v) > 0 {
			if errMsg, ok := v[0].(string); ok && errMsg == "NOT_FOUND" {
				return false, ErrBucketNotFound
			}
		}
		return false, fmt.Errorf("unexpected result format: %v", v)
	default:
		return false, fmt.Errorf("unexpected result type: %T", result)
	}
}

func (r *BucketRepository) Bucket(ctx context.Context, key string) (*model.Bucket, error) {
	result, err := r.client.HGetAll(bucketKey(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	if len(result) == 0 {
		return nil, ErrBucketNotFound
	}

	tokens, err := strconv.Atoi(result["tokens"])
	if err != nil {
		return nil, err
	}

	capacity, err := strconv.Atoi(result["capacity"])
	if err != nil {
		return nil, err
	}

	refilRate, err := strconv.Atoi(result["refil_rate"])
	if err != nil {
		return nil, err
	}

	lastRefill, err := strconv.ParseInt(result["last_refill"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &model.Bucket{
		Tokens:     tokens,
		Capacity:   capacity,
		RefilRate:  refilRate,
		LastRefill: time.Unix(lastRefill, 0),
	}, nil
}
