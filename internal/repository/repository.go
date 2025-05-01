package repository

import (
	"context"
	"errors"

	"github.com/vakhrushevk/cloudru/internal/repository/model"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
)

type BucketRepository interface {
	CreateBucket(ctx context.Context, key string, capacity int, refilRate int, tokens int) error
	Bucket(ctx context.Context, key string) (*model.Bucket, error)
	Decrease(ctx context.Context, key string) (bool, error)
	RefillAllBuckets(ctx context.Context) error
}
