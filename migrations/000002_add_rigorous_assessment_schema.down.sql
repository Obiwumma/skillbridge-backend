-- Migration: 000002_add_rigorous_assessment_schema.down.sql

DROP INDEX IF EXISTS idx_profiles_employability;
DROP INDEX IF EXISTS idx_assessment_sessions_ledger_id;
DROP INDEX IF EXISTS idx_assessment_ledgers_user_id;

DROP TABLE IF EXISTS assessment_sessions;
DROP TABLE IF EXISTS assessment_ledgers;

ALTER TABLE profiles 
DROP COLUMN IF EXISTS employability_score,
DROP COLUMN IF EXISTS highest_assessment_score,
DROP COLUMN IF EXISTS assessment_attempts,
DROP COLUMN IF EXISTS premium_vetting_passed;
