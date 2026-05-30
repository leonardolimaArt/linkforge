package integration

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/config"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/handler"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/origin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository/postgres"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/server"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/service"
)

const schemaSQL = `
CREATE TABLE short_links (
    id           uuid                     PRIMARY KEY,
    original_url varchar(2048)            NOT NULL,
    short_code   varchar(32)              NOT NULL UNIQUE,
    created_at   timestamp with time zone NOT NULL
);
`

var (
	testPool         *pgxpool.Pool
	testRedis        *redis.Client
	testRouter       *gin.Engine
	testKafkaBrokers []string
)

type noopOrigin struct{}

func (noopOrigin) Resolve(_ context.Context, _ string) (*domain.ShortLink, error) {
	return nil, origin.ErrOriginNotFound
}

func (noopOrigin) Close() error { return nil }

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgC, err := pgcontainer.Run(ctx,
		"postgres:16-alpine",
		pgcontainer.WithDatabase("linkforge_redirect"),
		pgcontainer.WithUsername("test"),
		pgcontainer.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start postgres: %v\n", err)
		os.Exit(1)
	}

	pgConnStr, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("failed to get postgres conn string: %v\n", err)
		os.Exit(1)
	}

	redisC, err := rediscontainer.Run(ctx, "redis:7-alpine")
	if err != nil {
		fmt.Printf("failed to start redis: %v\n", err)
		os.Exit(1)
	}

	rpC, err := redpanda.Run(ctx,
		"redpandadata/redpanda:v24.2.4",
		redpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		fmt.Printf("failed to start redpanda: %v\n", err)
		os.Exit(1)
	}

	seedBroker, err := rpC.KafkaSeedBroker(ctx)
	if err != nil {
		fmt.Printf("failed to get kafka seed broker: %v\n", err)
		os.Exit(1)
	}
	testKafkaBrokers = []string{seedBroker}

	redisHost, err := redisC.Host(ctx)
	if err != nil {
		fmt.Printf("failed to get redis hots: %v\n", err)
		os.Exit(1)
	}

	redisPort, err := redisC.MappedPort(ctx, "6379/tcp")
	if err != nil {
		fmt.Printf("failed to get redis port: %v\n", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, pgConnStr)
	if err != nil {
		fmt.Printf("failed to create pool %v\n", err)
		os.Exit(1)
	}

	if _, err := testPool.Exec(ctx, schemaSQL); err != nil {
		fmt.Printf("failed to create schema: %v\n", err)
		os.Exit(1)
	}

	testRedis = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
	})

	queries := postgres.New(testPool)
	repo := postgres.NewPgRepository(queries)
	redisCache := cache.NewRedisCache(testRedis)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewRedirectService(repo, redisCache, noopOrigin{}, testLogger)
	redirectHandler := handler.NewRedirectHandler(svc)
	healthHandler := handler.NewHealthHandler(testPool, testRedis)

	testCfg := &config.Config{
		CORSAllowedOrigins: "",
		RateLimitRPS:       10,
		RateLimitBurst:     5,
		AdminAPIKey:        "",
	}

	gin.SetMode(gin.TestMode)
	testRouter = server.New(testCfg, testLogger, redirectHandler, healthHandler)

	code := m.Run()

	testPool.Close()
	testRedis.Close()
	_ = pgC.Terminate(ctx)
	_ = redisC.Terminate(ctx)
	_ = rpC.Terminate(ctx)

	os.Exit(code)

}

func seedShortLink(t *testing.T, code, url string) {
	t.Helper()
	ctx := context.Background()
	_, err := testPool.Exec(ctx,
		`INSERT INTO short_links (id, original_url, short_code, created_at) VALUES ($1, $2, $3, $4)`,
		uuid.New(), url, code, time.Now(),
	)
	require.NoError(t, err)
}

func clearAll(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testPool.Exec(ctx, `TRUNCATE TABLE short_links`)
	require.NoError(t, err)
	require.NoError(t, testRedis.FlushDB(ctx).Err())
}

//Tests

func TestIntegration_Redirect_Sucess(t *testing.T) {
	clearAll(t)
	seedShortLink(t, "abc123", "https://google.com")

	req := httptest.NewRequest(http.MethodGet, "/r/abc123", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
}

func TestIntegration_Redirect_NotFound(t *testing.T) {
	clearAll(t)

	req := httptest.NewRequest(http.MethodGet, "/r/missing", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIntegration_Redirect_CachePopulatedAfterFirstHit(t *testing.T) {
	clearAll(t)
	seedShortLink(t, "abc", "https://google.com")

	req1 := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	w1 := httptest.NewRecorder()
	testRouter.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusFound, w1.Code)

	ctx := context.Background()
	cached, err := testRedis.Get(ctx, "redirect:abc").Result()
	require.NoError(t, err)
	assert.Equal(t, "https://google.com", cached)

	_, err = testPool.Exec(ctx, `DELETE FROM short_links WHERE short_code = $1`, "abc")
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	w2 := httptest.NewRecorder()
	testRouter.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusFound, w2.Code, "second request should be from cache")
}

func TestIntegration_RateLimiting_Triggers429(t *testing.T) {
	clearAll(t)
	seedShortLink(t, "rl_test", "http://google.com")

	statuse := make(map[int]int)
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodGet, "/r/rl_test", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		statuse[w.Code]++
	}

	assert.Greaterf(t, statuse[http.StatusTooManyRequests], 0, "expected at least one 429, got: %v", statuse)
	assert.Greaterf(t, statuse[http.StatusFound], 0, "expected at least one 302, got: %v", statuse)
}
