-- =============================================
-- Archery Tracker - Drop Database Schema
-- =============================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_stats_after_shot_change ON shots;
DROP TRIGGER IF EXISTS update_shots_updated_at ON shots;
DROP TRIGGER IF EXISTS update_sets_updated_at ON sets;
DROP TRIGGER IF EXISTS update_qualification_rounds_updated_at ON qualification_rounds;
DROP TRIGGER IF EXISTS update_target_faces_updated_at ON target_faces;

-- Drop functions
DROP FUNCTION IF EXISTS update_set_statistics();
DROP FUNCTION IF EXISTS calculate_set_grouping(UUID);
DROP FUNCTION IF EXISTS calculate_shot_score(DECIMAL, DECIMAL, UUID);
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (child first, parent last)
DROP TABLE IF EXISTS shots;
DROP TABLE IF EXISTS sets;
DROP TABLE IF EXISTS qualification_rounds;
DROP TABLE IF EXISTS target_faces;

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";