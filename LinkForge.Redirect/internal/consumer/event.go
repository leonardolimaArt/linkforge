package consumer

import (
	"time"

	"github.com/google/uuid"
)

type LinkCreatedEvent struct {
	SchemaVersion int       `json:"schema_version"`
	ID            uuid.UUID `json:"id"`
	ShortCode     string    `json:"short_code"`
	OriginalURL   string    `json:"original_url"`
	CreatedAt     time.Time `json:"created_at"`
}
