package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
)

type PgRepository struct {
	queries *Queries
}

func NewPgRepository(queries *Queries) *PgRepository {
	return &PgRepository{queries: queries}
}

func (r *PgRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.ShortLink, error) {
	row, err := r.queries.GetByShortCode(ctx, shortCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	return &domain.ShortLink{
		ID:          id,
		OriginalURL: row.OriginalUrl,
		ShortCode:   row.ShortCode,
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}
