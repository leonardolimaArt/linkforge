package origin

import (
	"context"
	"errors"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
)

var ErrOriginNotFound = errors.New("origin: short code not found")

type Client interface {
	Resolve(ctx context.Context, shortCode string) (*domain.ShortLink, error)
	Close() error
}
