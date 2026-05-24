package repository

import (
	"context"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
)

type Repository interface {
	GetByShortCode(ctx context.Context, shortCode string) (*domain.ShortLink, error)
	Upsert(ctx context.Context, link *domain.ShortLink) error
}
