package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository"
	"github.com/segmentio/kafka-go"
)

const (
	supportedSchemaVersion = 1
	cacheTTL               = time.Hour
)

type KafkaConsumer struct {
	reader *kafka.Reader
	repo   repository.Repository
	cache  cache.Cache
	logger *slog.Logger
}

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewKafkaConsumer(
	cfg Config,
	repo repository.Repository,
	cache cache.Cache,
	logger *slog.Logger,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
	})

	return &KafkaConsumer{
		reader: reader,
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (c *KafkaConsumer) Run(ctx context.Context) error {
	c.logger.Info("kafka consumer started",
		"topic", c.reader.Config().Topic,
		"group_id", c.reader.Config().GroupID,
	)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Info("kafka consumer stopping")
				return nil
			}
			c.logger.Error("fetch failed", "error", err)
			continue
		}

		if err := c.process(ctx, msg); err != nil {
			c.logger.Error("processing failed, will retry",
				"error", err,
				"offset", msg.Offset,
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit failed",
				"error", err,
				"offset", msg.Offset,
			)
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

func (c *KafkaConsumer) process(ctx context.Context, msg kafka.Message) error {
	var evt LinkCreatedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		c.logger.Warn("malformed message, skipping",
			"error", err,
			"offset", msg.Offset,
			"key", string(msg.Key),
		)
		return nil
	}

	if evt.SchemaVersion != supportedSchemaVersion {
		c.logger.Warn("unsupported schema version, skipping",
			"received", evt.SchemaVersion,
			"supported", supportedSchemaVersion,
			"offset", msg.Offset,
		)
		return nil
	}

	link := &domain.ShortLink{
		ID:          evt.ID,
		ShortCode:   evt.ShortCode,
		OriginalURL: evt.OriginalURL,
		CreatedAt:   evt.CreatedAt,
	}

	if err := c.repo.Upsert(ctx, link); err != nil {
		return err
	}

	if err := c.cache.Set(ctx, "redirect:"+evt.ShortCode, evt.OriginalURL, cacheTTL); err != nil {
		c.logger.Warn("cache set failed, ignoring",
			"error", err,
			"code", evt.ShortCode,
		)
	}

	c.logger.Info("event processed",
		"code", evt.ShortCode,
		"offset", msg.Offset,
	)
	return nil
}
