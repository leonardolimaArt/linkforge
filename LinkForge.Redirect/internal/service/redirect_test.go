package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/origin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockRepo struct {
	getFn       func(ctx context.Context, code string) (*domain.ShortLink, error)
	upsertFn    func(ctx context.Context, link *domain.ShortLink) error
	getCalls    int32
	upsertCalls int32
}

func (m *mockRepo) GetByShortCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	atomic.AddInt32(&m.getCalls, 1)
	return m.getFn(ctx, code)
}

func (m *mockRepo) Upsert(ctx context.Context, link *domain.ShortLink) error {
	atomic.AddInt32(&m.upsertCalls, 1)
	if m.upsertFn != nil {
		return m.upsertFn(ctx, link)
	}
	return nil
}

type mockCache struct {
	mu       sync.Mutex
	store    map[string]string
	getErr   error
	setErr   error
	setCalls int32
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.setErr != nil {
		return m.setErr
	}
	m.store[key] = value
	return nil
}

type mockOrigin struct {
	resolveFn    func(ctx context.Context, code string) (*domain.ShortLink, error)
	resolveCalls int32
}

func (m *mockOrigin) Resolve(ctx context.Context, code string) (*domain.ShortLink, error) {
	atomic.AddInt32(&m.resolveCalls, 1)
	if m.resolveFn != nil {
		return m.resolveFn(ctx, code)
	}
	return nil, origin.ErrOriginNotFound
}

func (m *mockOrigin) Close() error {
	return nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newServiceWith(repo *mockRepo, cache *mockCache, origin *mockOrigin) *RedirectService {
	return NewRedirectService(repo, cache, origin, newTestLogger())
}

// Tests
func TestResolve_L1_CacheHit(t *testing.T) {
	cache := newMockCache()
	cache.store["redirect:abc"] = "https://google.com"
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			t.Fatal("repo should not be called on cache hit")
			return nil, nil
		},
	}
	originMock := &mockOrigin{}
	svc := newServiceWith(repo, cache, originMock)
	url, err := svc.Resolve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(0), atomic.LoadInt32(&repo.getCalls))
	assert.Equal(t, int32(0), atomic.LoadInt32(&originMock.resolveCalls))
}

func TestResolve_L2_CacheMiss_PostgresHit(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return &domain.ShortLink{
				ID:          uuid.New(),
				ShortCode:   "abc",
				OriginalURL: "https://google.com",
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	originMock := &mockOrigin{}
	svc := newServiceWith(repo, cache, originMock)
	url, err := svc.Resolve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(1), atomic.LoadInt32(&repo.getCalls))
	assert.Equal(t, int32(0), atomic.LoadInt32(&originMock.resolveCalls))
	assert.Equal(t, "https://google.com", cache.store["redirect:abc"])
}

func TestResolve_L3_OriginHit_PopulatesPostgresAndCache(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, nil
		},
	}

	originID := uuid.New()
	originMock := &mockOrigin{
		resolveFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return &domain.ShortLink{
				ID:          originID,
				ShortCode:   code,
				OriginalURL: "https://google.com",
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	svc := newServiceWith(repo, cache, originMock)
	url, err := svc.Resolve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(1), atomic.LoadInt32(&originMock.resolveCalls))
	assert.Equal(t, int32(1), atomic.LoadInt32(&repo.upsertCalls))
	assert.Equal(t, "https://google.com", cache.store["redirect:abc"])
}

func TestResolve_L3_OriginNotFound_ReturnsErrNotFound(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, nil
		},
	}
	originMock := &mockOrigin{
		resolveFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, origin.ErrOriginNotFound
		},
	}

	svc := newServiceWith(repo, cache, originMock)
	_, err := svc.Resolve(context.Background(), "nope")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.Equal(t, int32(1), atomic.LoadInt32(&originMock.resolveCalls))
	assert.Equal(t, int32(0), atomic.LoadInt32(&repo.upsertCalls))
}

func TestResolve_L3_SkipsUpsertWhenOriginReturnsNilID(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, nil
		},
	}
	originMock := &mockOrigin{
		resolveFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return &domain.ShortLink{
				ID:          uuid.Nil,
				ShortCode:   code,
				OriginalURL: "https://google.com",
				CreatedAt:   time.Time{},
			}, nil
		},
	}
	svc := newServiceWith(repo, cache, originMock)
	url, err := svc.Resolve(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "https://google.com", url)
	assert.Equal(t, int32(0), atomic.LoadInt32(&repo.upsertCalls), "upsert should be skipped when ID is nil")
	assert.Equal(t, "https://google.com", cache.store["redirect:abc"])
}

func TestResolve_L3_OriginError_Propagates(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, nil
		},
	}
	originErr := errors.New("grpc: connection refused")
	originMock := &mockOrigin{
		resolveFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, originErr
		},
	}
	svc := newServiceWith(repo, cache, originMock)
	_, err := svc.Resolve(context.Background(), "abc")

	require.Error(t, err)
	assert.True(t, errors.Is(err, originErr))
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestResolve_DBError_DoesNotFallbackToOrigin(t *testing.T) {
	cache := newMockCache()
	dbErr := errors.New("connection refused")
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			return nil, dbErr
		},
	}
	originMock := &mockOrigin{}

	svc := newServiceWith(repo, cache, originMock)
	_, err := svc.Resolve(context.Background(), "abc")

	require.Error(t, err)
	assert.True(t, errors.Is(err, dbErr))
	assert.Equal(t, int32(0), atomic.LoadInt32(&originMock.resolveCalls), "origin should not be called when DB errors")
}

func TestResolve_SingleFlight_CoalescesConcurrentRequest(t *testing.T) {
	cache := newMockCache()
	repo := &mockRepo{
		getFn: func(ctx context.Context, code string) (*domain.ShortLink, error) {
			time.Sleep(50 * time.Millisecond)
			return &domain.ShortLink{
				ID:          uuid.New(),
				ShortCode:   "abcd123",
				OriginalURL: "https://googlr.com/abcd123",
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	originMock := &mockOrigin{}
	svc := newServiceWith(repo, cache, originMock)

	const concurent = 50
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

	assert.Equal(t, int32(1), atomic.LoadInt32(&repo.getCalls), "singleflight should coalesce concurrent requests")
}
