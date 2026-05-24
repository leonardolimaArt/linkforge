package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks
type mockRepo struct {
	getFn    func(ctx context.Context, code string) (*domain.ShortLink, error)
	getCalls int32
}

func (m *mockRepo) GetByShortCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	atomic.AddInt32(&m.getCalls, 1)
	return m.getFn(ctx, code)
}

type mockCache struct {
	store    map[string]string
	getErr   error
	setErr   error
	setCalls int32
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}

	v, ok := m.store[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	atomic.AddInt32(&m.setCalls, 1)
	if m.setErr != nil {
		return m.setErr
	}
	m.store[key] = value
	return nil
}

// Tests

func TestResolve_CacheHit(t *testing.T) {
	cache := newMockCache()
	cache.store["redirect:abc"] = "https://google.com"

	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			t.Fatal("repo should not be called on cache hit")
			return nil, nil
		},
	}

	svc := NewRedirectService(repo, cache)
	url, err := svc.Resolve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(0), atomic.LoadInt32(&repo.getCalls))
}

func TestResolve_CacheMiss_DBHit(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return &domain.ShortLink{
				ShortCode:   "abc",
				OriginalURL: "https://google.com",
			}, nil
		},
	}

	svc := NewRedirectService(repo, cache)
	url, err := svc.Resolve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(1), atomic.LoadInt32(&repo.getCalls))
	assert.Equal(t, "https://google.com", cache.store["redirect:abc"])
}

func TestResolve_NotFound(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, nil
		},
	}

	svc := NewRedirectService(repo, cache)
	_, err := svc.Resolve(context.Background(), "nope")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestResolve_DBError(t *testing.T) {
	cache := newMockCache()
	dbErr := errors.New("connection refused")

	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, dbErr
		},
	}

	svc := NewRedirectService(repo, cache)
	_, err := svc.Resolve(context.Background(), "abc")

	require.Error(t, err)
	assert.True(t, errors.Is(err, dbErr))
}

func TestResolve_SingleFlight_CoalescesConcurrentRequest(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			time.Sleep(50 * time.Millisecond)
			return &domain.ShortLink{
				ShortCode:   "abcd123",
				OriginalURL: "https://google.com/abcd123",
			}, nil
		},
	}

	svc := NewRedirectService(repo, cache)

	const concurent = 5000000
	done := make(chan error, concurent)
	for i := 0; i < concurent; i++ {
		go func() {
			_, err := svc.Resolve(context.Background(), "abcd123")
			done <- err
		}()
	}

	for i := 0; i < concurent; i++ {
		require.NoError(t, <-done)
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&repo.getCalls))
}
