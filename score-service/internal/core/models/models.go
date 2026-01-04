package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateQualificationRoundRequest struct {
	RoundType    string     `json:"round_type"` // training, qualification, practice, warmup
	Name         string     `json:"name"`
	Distance     int        `json:"distance"` // in meters
	TotalSets    int        `json:"total_sets"`
	ShotsPerSet  int        `json:"shots_per_set"`
	TargetFaceID uuid.UUID  `json:"target_face_id"`
	Notes        string     `json:"notes,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
}

type CreateSetRequest struct {
	SetNumber     int       `json:"set_number"`
	MaxShots      *int      `json:"max_shots,omitempty"`
	ParentRoundId uuid.UUID `json:"parent_round_id"`
}

type CreateShotRequest struct {
	X     float64 `json:"x"` // horizontal offset in mm
	Y     float64 `json:"y"` // vertical offset in mm
	Score int8    `json:"score"`
	Notes string  `json:"notes,omitempty"`
}

type CreateShotsBatchRequest struct {
	Shots []CreateShotRequest `json:"shots"`
}

type UpdateShotRequest struct {
	X     *float64 `json:"x,omitempty"`
	Y     *float64 `json:"y,omitempty"`
	Notes *string  `json:"notes,omitempty"`
}

type QualificationRoundResponse struct {
	ID             uuid.UUID     `json:"id"`
	ExternalUserID string        `json:"external_user_id"`
	RoundType      string        `json:"round_type"`
	Name           string        `json:"name"`
	Distance       int           `json:"distance"`
	TotalSets      int           `json:"total_sets"`
	ShotsPerSet    int           `json:"shots_per_set"`
	TotalScore     int           `json:"total_score"`
	AverageScore   float64       `json:"average_score"`
	CompletedSets  int           `json:"completed_sets"`
	StartTime      *time.Time    `json:"start_time,omitempty"`
	EndTime        *time.Time    `json:"end_time,omitempty"`
	Notes          string        `json:"notes,omitempty"`
	TargetFaceID   uuid.UUID     `json:"target_face_id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Sets           []SetResponse `json:"sets,omitempty"`
}

type SetResponse struct {
	ID               uuid.UUID      `json:"id"`
	SetNumber        int            `json:"set_number"`
	MaxShots         int            `json:"max_shots"`
	TotalScore       int            `json:"total_score"`
	AverageScore     float64        `json:"average_score"`
	ShotsCount       int            `json:"shots_count"`
	TenCount         int            `json:"ten_count"`
	XCount           int            `json:"x_count"`
	MissCount        int            `json:"miss_count"`
	GroupingDiameter *float64       `json:"grouping_diameter,omitempty"`
	GroupingCenterX  *float64       `json:"grouping_center_x,omitempty"`
	GroupingCenterY  *float64       `json:"grouping_center_y,omitempty"`
	ParentRoundID    uuid.UUID      `json:"parent_round_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Shots            []ShotResponse `json:"shots,omitempty"`
}

type ShotResponse struct {
	ID                 uuid.UUID `json:"id"`
	X                  float64   `json:"x"`
	Y                  float64   `json:"y"`
	Score              int       `json:"score"`
	DistanceFromCenter float64   `json:"distance_from_center"`
	IsTen              bool      `json:"is_ten"`
	IsX                bool      `json:"is_x"`
	IsMiss             bool      `json:"is_miss"`
	Notes              string    `json:"notes,omitempty"`
	SetID              uuid.UUID `json:"set_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
