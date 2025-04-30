package repository

import (
	"context"

	"github.com/vakhrushevk/cloudru/internal/repository/model"
)

type ClientRepository interface {
	CreateClient(ctx context.Context, client *model.Client)
	ClientSettings(ctx context.Context, client *model.Client)
}

type TokenRepository interface {
	GetTokens(ctx context.Context, url string) (int, error)
}
