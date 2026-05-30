package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/google/uuid"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/origin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository"
)

var ErrNotFound = errors.New("short link not found")

const cacheKeyPrefix = "redirect:"
const cacheTTL = time.Hour

type RedirectService struct {
	repo   repository.Repository
	cache  cache.Cache
	sf     singleflight.Group
	logger *slog.Logger
	origin origin.Client
}

func NewRedirectService(repo repository.Repository, cache cache.Cache, origin origin.Client, logger *slog.Logger) *RedirectService {
	return &RedirectService{
		repo:   repo,
		cache:  cache,
		origin: origin,
		logger: logger,
	}
}

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (string, error) {
	url, err := s.cache.Get(ctx, cacheKeyPrefix+shortCode)
	if err == nil && url != "" {
		s.logger.Info("L1 cache hit", "short_code", shortCode)
		return url, nil
	}

	v, err, _ := s.sf.Do(shortCode, func() (any, error) {
		link, err := s.repo.GetByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if link != nil {
			s.logger.Info("L2 postgres hit", "short_code", shortCode)
			_ = s.cache.Set(ctx, cacheKeyPrefix+shortCode, link.OriginalURL, cacheTTL)
			return link.OriginalURL, nil
		}
		s.logger.Info("L3 grpc fallback", "short_code", shortCode)
		originLink, err := s.origin.Resolve(ctx, shortCode)
		if errors.Is(err, origin.ErrOriginNotFound) {
			return "", ErrNotFound
		}

		if err != nil {
			s.logger.Error("origin fallback failed", "error", err, "short_code", shortCode)
			return "", err
		}

		if originLink.ID != uuid.Nil {
			if upsertErr := s.repo.Upsert(ctx, originLink); upsertErr != nil {
				s.logger.Warn("upsert after origin fallback failed",
					"error", upsertErr,
					"short_code", shortCode)
			}
		}

		_ = s.cache.Set(ctx, cacheKeyPrefix+shortCode, originLink.OriginalURL, cacheTTL)
		return originLink.OriginalURL, nil
	})

	if err != nil {
		return "", err
	}

	return v.(string), nil

}
