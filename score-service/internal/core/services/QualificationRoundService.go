package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"archy/scores/internal/core/models"
	"archy/scores/internal/db"
)

type QualificationRoundService struct {
	queries *db.Queries
}

func NewQualificationRoundService(queries *db.Queries) *QualificationRoundService {
	return &QualificationRoundService{queries: queries}
}

func (s *QualificationRoundService) CreateQualificationRound(
	ctx context.Context,
	externalUserID string,
	req models.CreateQualificationRoundRequest,
) (*db.QualificationRound, error) {
	// Подготавливаем параметры для запроса
	params := db.CreateQualificationRoundParams{
		ExternalUserID: externalUserID,
		RoundType:      req.RoundType,
		Name:           req.Name,
		Distance:       int32(req.Distance),
		TotalSets:      int32(req.TotalSets),
		ShotsPerSet:    int32(req.ShotsPerSet),
		TargetFaceID:   req.TargetFaceID,
		Notes:          pgtype.Text{String: req.Notes, Valid: true},
	}

	// Время начала
	if req.StartTime != nil {
		params.StartTime = pgtype.Timestamptz{
			Time:  *req.StartTime,
			Valid: true,
		}
	} else {
		params.StartTime = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}
	res, err := s.queries.CreateQualificationRound(ctx, params)

	return &res, err
}

func (s *QualificationRoundService) GetQualificationRound(
	ctx context.Context,
	externalUserID string,
	roundID uuid.UUID,
) (*db.QualificationRound, error) {
	round, err := s.queries.GetQualificationRound(ctx, roundID)
	if err != nil {
		return nil, err
	}

	if round.ExternalUserID != externalUserID {
		return nil, echo.NewHTTPError(404, "Round not found")
	}

	return &round, nil
}

func (s *QualificationRoundService) GetQualificationRoundsForUser(
	ctx context.Context,
	externalUserID string,
) ([]db.QualificationRound, error) {
	rounds, err := s.queries.GetQualificationRoundsForUser(ctx, externalUserID)
	if err != nil {
		return nil, err
	}

	if rounds == nil {
		return []db.QualificationRound{}, nil
	}

	return rounds, nil
}
