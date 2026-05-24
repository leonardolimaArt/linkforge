package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShortLink struct {
	ID	uuid.UUID
	OriginalURL	string
	ShortCode	string
	CreatedAt	time.Time
}