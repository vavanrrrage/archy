package services

import (
	"archy/scores/internal/core/models"
	"archy/scores/internal/db"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SetService struct {
	queries *db.Queries
}

func NewSetService(queries *db.Queries) *SetService {
	return &SetService{queries: queries}
}

func (s *SetService) CreateSet(
	ctx context.Context,
	req models.CreateSetRequest,
) (*db.Set, error) {
	params := db.CreateSetParams{
		SetNumber:     int32(req.SetNumber),
		MaxShots:      int32(*req.MaxShots),
		ParentRoundID: pgtype.UUID{Bytes: req.ParentRoundId},
	}

	set, err := s.queries.CreateSet(
		ctx,
		params,
	)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

func (s *SetService) GetSet(ctx context.Context, id uuid.UUID) (*db.Set, error) {
	set, err := s.queries.GetSet(ctx, id)
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func (s *SetService) GetSetShots(ctx context.Context, id uuid.UUID) ([]db.Shot, error) {
	shots, err := s.queries.GetShotsBySet(ctx, id)
	if err != nil {
		return nil, err
	}
	return shots, nil
}

func (s *SetService) GetSetsForQualificationRound(ctx context.Context, roundId pgtype.UUID) ([]db.Set, error) {
	sets, err := s.queries.GetSetsForQualificationRound(ctx, roundId)

	if err != nil {
		return nil, err
	}

	return sets, err
}
