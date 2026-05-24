package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository"
)

var ErrNotFound = errors.New("short link not found")

type RedirectService struct {
	repo repository.Repository
	cache cache.Cache
	sf singleflight.Group
}

func NewRedirectService(repo repository.Repository, cache cache.Cache) *RedirectService {
	return &RedirectService{
		repo: repo,
		cache: cache,
	}
}

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (string, error) {
	url, err := s.cache.Get(ctx, "redirect:"+shortCode)
	if err == nil && url != "" {
		return url, nil
	}

	v, err, _ := s.sf.Do(shortCode, func() (any, error) {
		link, err := s.repo.GetByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if link == nil {
			return "", ErrNotFound
		}
		_ = s.cache.Set(ctx, "redirect:"+shortCode, link.OriginalURL, time.Hour)
		return link.OriginalURL, nil
	})

	if err != nil {
		return "", err
	}

	return v.(string), nil

}