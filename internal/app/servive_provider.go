package app

import (
	"context"
	"log"

	"github.com/go-redis/redis"
	"github.com/vakhrushevk/cloudru/internal/balancer"
	"github.com/vakhrushevk/cloudru/internal/config"
	ratelimit "github.com/vakhrushevk/cloudru/internal/rateLimit"
	"github.com/vakhrushevk/cloudru/internal/repository"
	"github.com/vakhrushevk/cloudru/internal/repository/redisRepository"
	"github.com/vakhrushevk/cloudru/pkg/logger"
)

type serviceProvider struct {
	redisClient      *redis.Client
	limiter          *ratelimit.Limiter
	balancer         balancer.Balancer
	bucketRepository repository.BucketRepository
	config           *config.Config
}

// NewServiceProvider создает новый сервис-провайдер
func newServiceProvider(_ context.Context) (*serviceProvider, error) {
	s := &serviceProvider{}
	s.InitLogger()
	return s, nil
}

// InitLogger инициализирует логгер
func (s *serviceProvider) InitLogger() {
	logger.Init(&s.Config().LoggerConfig)
}

// Config возвращает конфигурацию или загружает ее
func (s *serviceProvider) Config() *config.Config {
	if s.config == nil {
		cfg, err := config.LoadConfig(globalConfigPath)
		if err != nil {
			log.Fatal("error loading config:", err)
		}
		s.config = cfg
	}

	return s.config
}

// RedisClient создает новый клиент Redis или возвращает существующий
func (s *serviceProvider) RedisClient(_ context.Context) *redis.Client {
	if s.redisClient == nil {
		s.redisClient = redis.NewClient(&redis.Options{
			Addr:     s.Config().RedisConfig.Addr,
			Password: s.Config().RedisConfig.Password,
			DB:       s.Config().RedisConfig.DB,
		})
		_, err := s.redisClient.Ping().Result()
		if err != nil {
			log.Fatal("error connecting to redis:", err)
		}
	}

	return s.redisClient
}

// BucketRepository создает новый репозиторий бакетов или возвращает существующий
func (s *serviceProvider) BucketRepository(ctx context.Context) repository.BucketRepository {
	if s.bucketRepository == nil {
		bucketRepo, err := redisRepository.NewRedisRepository(s.RedisClient(ctx))
		if err != nil {
			log.Fatal("error creating bucket repository:", err)
		}
		s.bucketRepository = bucketRepo
	}

	return s.bucketRepository
}

// Limiter создает новый лимитер или возвращает существующий
func (s *serviceProvider) Limiter(ctx context.Context) *ratelimit.Limiter {
	if s.limiter == nil {
		s.limiter = ratelimit.NewLimiter(ctx, s.BucketRepository(ctx), s.Config().BucketConfig)
	}
	return s.limiter
}

// Balancer создает новый балансер или возвращает существующий
func (s *serviceProvider) Balancer(ctx context.Context) balancer.Balancer {
	if s.balancer == nil {
		balance, err := balancer.New(ctx, s.Config().BalancerConfig, s.Config().RetryConfig)
		balancer.CheckAndUpdate(*s.Config(), balance)
		if err != nil {
			log.Fatal("error creating balancer:", err)
		}
		s.balancer = balance
	}
	return s.balancer
}
