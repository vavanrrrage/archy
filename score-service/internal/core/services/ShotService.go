package services

import (
	"archy/scores/internal/core/models"
	"archy/scores/internal/db"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ShotService struct {
	queries *db.Queries
}

func NewShotService(queries *db.Queries) *ShotService {
	return &ShotService{queries: queries}
}

func (s *ShotService) CreateShot(
	ctx context.Context,
	setId uuid.UUID,
	shot models.CreateShotRequest,
) (*db.Shot, error) {
	var x pgtype.Numeric
	var y pgtype.Numeric

	errX := x.Scan(shot.X)
	if errX != nil {
		return nil, errX
	}
	errY := y.Scan(shot.Y)
	if errY != nil {
		return nil, errY
	}

	var text pgtype.Text
	errT := text.Scan(shot.Notes)
	if errT != nil {
		return nil, errT
	}

	params := db.CreateShotParams{
		X:     x,
		Y:     y,
		Notes: text,
		SetID: setId,
	}
	sh, err := s.queries.CreateShot(ctx, params)
	if err != nil {
		return nil, err
	}
	return &sh, nil
}

func (s *ShotService) GetShotsBySet(ctx context.Context, setId uuid.UUID) ([]db.Shot, error) {
	sh, err := s.queries.GetShotsBySet(ctx, setId)
	if err != nil {
		return nil, err
	}
	return sh, nil
}

func (s *ShotService) GetShot(ctx context.Context, shotId uuid.UUID) (*db.Shot, error) {
	sh, err := s.queries.GetShot(ctx, shotId)
	if err != nil {
		return nil, err
	}
	return &sh, nil
}
