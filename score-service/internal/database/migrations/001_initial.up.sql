-- =============================================
-- Archery Tracker - Initial Database Schema
-- Version: 1.0
-- Description: Initial schema for archery tracking system
--              Users are identified by external_id from auth service
-- =============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================
-- TARGET FACES
-- =============================================
-- Stores different target configurations (WA, FITA, etc.)
CREATE TABLE target_faces (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              name VARCHAR(100) NOT NULL UNIQUE,
                              standard VARCHAR(50) NOT NULL,
                              total_diameter INTEGER NOT NULL,          -- in mm
                              scoring_diameter INTEGER NOT NULL,        -- in mm
                              zones_config JSONB NOT NULL,
                              max_score INTEGER NOT NULL,
                              has_x BOOLEAN NOT NULL DEFAULT false,
                              description TEXT,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              deleted_at TIMESTAMPTZ
);

-- Indexes for target_faces
CREATE INDEX idx_target_faces_standard ON target_faces(standard);
CREATE INDEX idx_target_faces_deleted ON target_faces(deleted_at) WHERE deleted_at IS NULL;
COMMENT ON TABLE target_faces IS 'Target face configurations (WA 122cm, WA 80cm, etc.)';

-- =============================================
-- QUALIFICATION ROUNDS / TRAINING SESSIONS
-- =============================================
-- Main entity representing a training session or qualification round
CREATE TABLE qualification_rounds (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      external_user_id VARCHAR(255) NOT NULL,   -- ID from auth service
                                      round_type VARCHAR(50) NOT NULL,          -- 'training', 'qualification', 'practice', 'warmup'
                                      name VARCHAR(100) NOT NULL,
                                      distance INTEGER NOT NULL,                -- in meters
                                      total_sets INTEGER NOT NULL,
                                      shots_per_set INTEGER NOT NULL,
                                      total_score INTEGER NOT NULL DEFAULT 0,
                                      average_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
                                      completed_sets INTEGER NOT NULL DEFAULT 0,
                                      start_time TIMESTAMPTZ,
                                      end_time TIMESTAMPTZ,
                                      notes TEXT,
                                      target_face_id UUID NOT NULL REFERENCES target_faces(id),
                                      competition_id UUID,                      -- optional link to external competition
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      deleted_at TIMESTAMPTZ
);

-- Indexes for qualification_rounds
CREATE INDEX idx_qualification_rounds_user ON qualification_rounds(external_user_id);
CREATE INDEX idx_qualification_rounds_type ON qualification_rounds(round_type);
CREATE INDEX idx_qualification_rounds_distance ON qualification_rounds(distance);
CREATE INDEX idx_qualification_rounds_target ON qualification_rounds(target_face_id);
CREATE INDEX idx_qualification_rounds_created ON qualification_rounds(created_at DESC);
CREATE INDEX idx_qualification_rounds_deleted ON qualification_rounds(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_qualification_rounds_user_created ON qualification_rounds(external_user_id, created_at DESC);
COMMENT ON TABLE qualification_rounds IS 'Training sessions or qualification rounds';

-- =============================================
-- SETS
-- =============================================
-- A series of shots within a round
CREATE TABLE sets (
                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                      set_number INTEGER NOT NULL,              -- 1, 2, 3... within a round
                      max_shots INTEGER NOT NULL DEFAULT 6,
                      total_score INTEGER NOT NULL DEFAULT 0,
                      average_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
                      shots_count INTEGER NOT NULL DEFAULT 0,
                      ten_count INTEGER NOT NULL DEFAULT 0,
                      x_count INTEGER NOT NULL DEFAULT 0,
                      miss_count INTEGER NOT NULL DEFAULT 0,
                      grouping_diameter DECIMAL(5,2),           -- in mm
                      grouping_center_x DECIMAL(5,2),          -- in mm from center
                      grouping_center_y DECIMAL(5,2),          -- in mm from center
                      parent_round_id UUID REFERENCES qualification_rounds(id) ON DELETE CASCADE,
                      parent_match_id UUID,                     -- for future match play support
                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                      deleted_at TIMESTAMPTZ,
                      CONSTRAINT check_parent CHECK (
                          (parent_round_id IS NOT NULL AND parent_match_id IS NULL) OR
                          (parent_round_id IS NULL AND parent_match_id IS NOT NULL)
                      )
);

-- Indexes for sets
CREATE INDEX idx_sets_parent_round ON sets(parent_round_id);
CREATE INDEX idx_sets_parent_match ON sets(parent_match_id);
CREATE INDEX idx_sets_set_number ON sets(set_number);
CREATE INDEX idx_sets_deleted ON sets(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_sets_round_set_number ON sets(parent_round_id, set_number);
COMMENT ON TABLE sets IS 'Series of shots (typically 3 or 6 arrows)';

-- =============================================
-- SHOTS
-- =============================================
-- Individual shot records with coordinates
CREATE TABLE shots (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       x DECIMAL(5,2) NOT NULL,                 -- horizontal offset from center in mm
                       y DECIMAL(5,2) NOT NULL,                 -- vertical offset from center in mm
                       score INTEGER NOT NULL,                   -- calculated score (0-10)
                       distance_from_center DECIMAL(5,2) NOT NULL, -- in mm
                       is_ten BOOLEAN NOT NULL DEFAULT false,
                       is_x BOOLEAN NOT NULL DEFAULT false,
                       is_miss BOOLEAN NOT NULL DEFAULT false,
                       notes TEXT,
                       set_id UUID NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       deleted_at TIMESTAMPTZ
);

-- Indexes for shots
CREATE INDEX idx_shots_score ON shots(score);
CREATE INDEX idx_shots_is_ten ON shots(is_ten);
CREATE INDEX idx_shots_is_x ON shots(is_x);
CREATE INDEX idx_shots_is_miss ON shots(is_miss);
CREATE INDEX idx_shots_set ON shots(set_id);
CREATE INDEX idx_shots_created ON shots(created_at);
CREATE INDEX idx_shots_deleted ON shots(deleted_at) WHERE deleted_at IS NULL;
COMMENT ON TABLE shots IS 'Individual shot records with coordinates and score';

-- =============================================
-- FUNCTIONS
-- =============================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate score from coordinates and target face
CREATE OR REPLACE FUNCTION calculate_shot_score(
    p_x DECIMAL,
    p_y DECIMAL,
    p_target_face_id UUID
) RETURNS INTEGER AS $$
DECLARE
v_distance DECIMAL;
    v_zones JSONB;
    v_zone JSONB;
    v_score INTEGER := 0;
BEGIN
    -- Calculate Euclidean distance from center (in mm)
    v_distance := SQRT(POWER(p_x, 2) + POWER(p_y, 2));

    -- Get zones configuration
SELECT zones_config INTO v_zones
FROM target_faces
WHERE id = p_target_face_id AND deleted_at IS NULL;

-- Find the smallest zone that contains the distance
FOR v_zone IN SELECT * FROM jsonb_array_elements(v_zones)
              ORDER BY (value->>'radius')::DECIMAL ASC
    LOOP
        IF v_distance <= (v_zone->>'radius')::DECIMAL THEN
            v_score := (v_zone->>'score')::INTEGER;
EXIT;
END IF;
END LOOP;

RETURN v_score;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to calculate grouping statistics for a set
CREATE OR REPLACE FUNCTION calculate_set_grouping(p_set_id UUID)
RETURNS TABLE (
    grouping_diameter DECIMAL,
    grouping_center_x DECIMAL,
    grouping_center_y DECIMAL
) AS $$
BEGIN
RETURN QUERY
    WITH shot_stats AS (
        SELECT
            STDDEV_SAMP(x) as std_x,
            STDDEV_SAMP(y) as std_y,
            AVG(x) as avg_x,
            AVG(y) as avg_y,
            COUNT(*) as count
        FROM shots
        WHERE set_id = p_set_id AND deleted_at IS NULL
    )
SELECT
    CASE
        WHEN count >= 2 THEN SQRT(POWER(std_x, 2) + POWER(std_y, 2)) * 2
        ELSE NULL
        END as grouping_diameter,
    avg_x as grouping_center_x,
    avg_y as grouping_center_y
FROM shot_stats;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- TRIGGERS
-- =============================================

-- Update triggers for all tables
CREATE TRIGGER update_target_faces_updated_at
    BEFORE UPDATE ON target_faces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_qualification_rounds_updated_at
    BEFORE UPDATE ON qualification_rounds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sets_updated_at
    BEFORE UPDATE ON sets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shots_updated_at
    BEFORE UPDATE ON shots
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update set statistics when shots change
CREATE OR REPLACE FUNCTION update_set_statistics()
RETURNS TRIGGER AS $$
BEGIN
    -- Update statistics for the affected set
UPDATE sets s
SET
    shots_count = ss.shots_count,
    total_score = ss.total_score,
    average_score = ss.average_score,
    ten_count = ss.ten_count,
    x_count = ss.x_count,
    miss_count = ss.miss_count,
    grouping_diameter = g.grouping_diameter,
    grouping_center_x = g.grouping_center_x,
    grouping_center_y = g.grouping_center_y,
    updated_at = NOW()
    FROM (
        SELECT
            set_id,
            COUNT(*) as shots_count,
            SUM(score) as total_score,
            AVG(score) as average_score,
            COUNT(CASE WHEN is_ten THEN 1 END) as ten_count,
            COUNT(CASE WHEN is_x THEN 1 END) as x_count,
            COUNT(CASE WHEN is_miss THEN 1 END) as miss_count
        FROM shots
        WHERE set_id = COALESCE(NEW.set_id, OLD.set_id)
          AND deleted_at IS NULL
        GROUP BY set_id
    ) ss
    LEFT JOIN LATERAL calculate_set_grouping(COALESCE(NEW.set_id, OLD.set_id)) g ON true
WHERE s.id = COALESCE(NEW.set_id, OLD.set_id);

-- Update parent round statistics
UPDATE qualification_rounds qr
SET
    total_score = rs.total_score,
    average_score = rs.average_score,
    completed_sets = rs.completed_sets,
    updated_at = NOW()
    FROM (
        SELECT
            parent_round_id,
            SUM(total_score) as total_score,
            AVG(average_score) as average_score,
            COUNT(*) as completed_sets
        FROM sets
        WHERE parent_round_id = (
            SELECT parent_round_id
            FROM sets
            WHERE id = COALESCE(NEW.set_id, OLD.set_id)
        ) AND deleted_at IS NULL
        GROUP BY parent_round_id
    ) rs
WHERE qr.id = rs.parent_round_id;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_stats_after_shot_change
    AFTER INSERT OR UPDATE OR DELETE ON shots
    FOR EACH ROW EXECUTE FUNCTION update_set_statistics();

-- =============================================
-- DEFAULT DATA
-- =============================================

-- Insert common target faces
INSERT INTO target_faces (name, standard, total_diameter, scoring_diameter, zones_config, max_score, has_x, description) VALUES
                                                                                                                             (
                                                                                                                                 'WA 122cm 10-zone',
                                                                                                                                 'WA',
                                                                                                                                 1220,
                                                                                                                                 800,
                                                                                                                                 '[
                                                                                                                                   {"score": 10, "radius": 61.0, "color": "gold", "hasInnerRing": true, "innerRadius": 30.5},
                                                                                                                                   {"score": 9, "radius": 122.0, "color": "red"},
                                                                                                                                   {"score": 8, "radius": 183.0, "color": "red"},
                                                                                                                                   {"score": 7, "radius": 244.0, "color": "blue"},
                                                                                                                                   {"score": 6, "radius": 305.0, "color": "blue"},
                                                                                                                                   {"score": 5, "radius": 366.0, "color": "black"},
                                                                                                                                   {"score": 4, "radius": 427.0, "color": "black"},
                                                                                                                                   {"score": 3, "radius": 488.0, "color": "white"},
                                                                                                                                   {"score": 2, "radius": 549.0, "color": "white"},
                                                                                                                                   {"score": 1, "radius": 610.0, "color": "white"}
                                                                                                                                 ]'::jsonb,
                                                                                                                                 10,
                                                                                                                                 true,
                                                                                                                                 'World Archery 122cm target face (70m, 60m, 50m)'
                                                                                                                             ),
                                                                                                                             (
                                                                                                                                 'WA 80cm 10-zone',
                                                                                                                                 'WA',
                                                                                                                                 800,
                                                                                                                                 400,
                                                                                                                                 '[
                                                                                                                                   {"score": 10, "radius": 40.0, "color": "gold", "hasInnerRing": true, "innerRadius": 20.0},
                                                                                                                                   {"score": 9, "radius": 80.0, "color": "red"},
                                                                                                                                   {"score": 8, "radius": 120.0, "color": "red"},
                                                                                                                                   {"score": 7, "radius": 160.0, "color": "blue"},
                                                                                                                                   {"score": 6, "radius": 200.0, "color": "blue"},
                                                                                                                                   {"score": 5, "radius": 240.0, "color": "black"},
                                                                                                                                   {"score": 4, "radius": 280.0, "color": "black"},
                                                                                                                                   {"score": 3, "radius": 320.0, "color": "white"},
                                                                                                                                   {"score": 2, "radius": 360.0, "color": "white"},
                                                                                                                                   {"score": 1, "radius": 400.0, "color": "white"}
                                                                                                                                 ]'::jsonb,
                                                                                                                                 10,
                                                                                                                                 true,
                                                                                                                                 'World Archery 80cm target face (30m, 18m indoor)'
                                                                                                                             ),
                                                                                                                             (
                                                                                                                                 '3-Spot Vertical',
                                                                                                                                 'NFAA',
                                                                                                                                 400,
                                                                                                                                 400,
                                                                                                                                 '[
                                                                                                                                   {"score": 5, "radius": 40.0, "color": "white"},
                                                                                                                                   {"score": 4, "radius": 80.0, "color": "black"},
                                                                                                                                   {"score": 3, "radius": 120.0, "color": "black"},
                                                                                                                                   {"score": 2, "radius": 160.0, "color": "white"},
                                                                                                                                   {"score": 1, "radius": 200.0, "color": "white"}
                                                                                                                                 ]'::jsonb,
                                                                                                                                 5,
                                                                                                                                 false,
                                                                                                                                 'NFAA 3-spot vertical target (indoor)'
                                                                                                                             );

-- =============================================
-- COMMENTS
-- =============================================
COMMENT ON COLUMN qualification_rounds.external_user_id IS 'User ID from external authentication service (JWT claim: user_id or sub)';
COMMENT ON COLUMN shots.x IS 'Horizontal offset from center in mm. Positive = right, Negative = left';
COMMENT ON COLUMN shots.y IS 'Vertical offset from center in mm. Positive = up, Negative = down';
COMMENT ON COLUMN sets.grouping_diameter IS 'Diameter of arrow grouping in mm (2 * standard deviation)';