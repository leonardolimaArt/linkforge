package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/consumer"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository/postgres"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kafkaTopic = "linkforge.links.created"

func runConsumer(t *testing.T, ctx context.Context) {
	t.Helper()

	queries := postgres.New(testPool)
	repo := postgres.NewPgRepository(queries)
	redisCache := cache.NewRedisCache(testRedis)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	c := consumer.NewKafkaConsumer(
		consumer.Config{
			Brokers: testKafkaBrokers,
			Topic:   kafkaTopic,
			GroupID: "test-group-" + uuid.NewString(),
		},
		repo,
		redisCache,
		logger,
	)

	go func() {
		_ = c.Run(ctx)
	}()

	t.Cleanup(func() {
		_ = c.Close()
	})
}

func publishKafkaMessage(t *testing.T, key string, payload []byte) {
	t.Helper()

	w := &kafka.Writer{
		Addr:                   kafka.TCP(testKafkaBrokers...),
		Topic:                  kafkaTopic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}
	defer w.Close()

	err := w.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: payload,
	})
	require.NoError(t, err)
}

func publishEvent(t *testing.T, event consumer.LinkCreatedEvent) {
	t.Helper()
	body, err := json.Marshal(event)
	require.NoError(t, err)
	publishKafkaMessage(t, event.ShortCode, body)
}

func TestIntegration_KafkaConsumer_ProcessesValidEvent(t *testing.T) {
	clearAll(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runConsumer(t, ctx)

	id := uuid.New()
	publishEvent(t, consumer.LinkCreatedEvent{
		SchemaVersion: 1,
		ID:            id,
		ShortCode:     "int_ok01",
		OriginalURL:   "https://google.com/integration",
		CreatedAt:     time.Now().UTC().Truncate(time.Second),
	})

	require.Eventually(t, func() bool {
		queries := postgres.New(testPool)
		repo := postgres.NewPgRepository(queries)
		link, err := repo.GetByShortCode(context.Background(), "int_ok01")
		return err == nil && link != nil
	}, 15*time.Second, 200*time.Millisecond, "consumer never processed the event")

	queries := postgres.New(testPool)
	repo := postgres.NewPgRepository(queries)
	link, err := repo.GetByShortCode(context.Background(), "int_ok01")
	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, id, link.ID)
	assert.Equal(t, "https://google.com/integration", link.OriginalURL)

	cached, err := testRedis.Get(context.Background(), "redirect:int_ok01").Result()
	require.NoError(t, err)
	assert.Equal(t, "https://google.com/integration", cached)
}

func TestIntegration_KafkaConsumer_SkipsMalformedMessage(t *testing.T) {
	clearAll(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runConsumer(t, ctx)

	publishKafkaMessage(t, "poison", []byte("not valid json {{{"))

	publishEvent(t, consumer.LinkCreatedEvent{
		SchemaVersion: 1,
		ID:            uuid.New(),
		ShortCode:     "after_poison",
		OriginalURL:   "https://google.com/recovered",
		CreatedAt:     time.Now().UTC().Truncate(time.Second),
	})

	require.Eventually(t, func() bool {
		queries := postgres.New(testPool)
		repo := postgres.NewPgRepository(queries)
		link, err := repo.GetByShortCode(context.Background(), "after_poison")
		return err == nil && link != nil
	}, 15*time.Second, 200*time.Millisecond, "consumer got stuck on poison message")
}

func TestIntegration_KafkaConsumer_IsIdempotent(t *testing.T) {
	clearAll(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runConsumer(t, ctx)

	id := uuid.New()
	event := consumer.LinkCreatedEvent{
		SchemaVersion: 1,
		ID:            id,
		ShortCode:     "idempot1",
		OriginalURL:   "https://google.com/idempotent",
		CreatedAt:     time.Now().UTC().Truncate(time.Second),
	}

	publishEvent(t, event)
	publishEvent(t, event)
	publishEvent(t, event)

	require.Eventually(t, func() bool {
		queries := postgres.New(testPool)
		repo := postgres.NewPgRepository(queries)
		link, _ := repo.GetByShortCode(context.Background(), "idempot1")
		return link != nil
	}, 15*time.Second, 200*time.Millisecond)

	time.Sleep(2 * time.Second)

	var count int
	err := testPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM short_links WHERE short_code = $1`,
		"idempot1",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
