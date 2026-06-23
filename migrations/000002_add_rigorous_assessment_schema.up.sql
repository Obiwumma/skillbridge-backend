-- Migration: 000002_add_rigorous_assessment_schema.up.sql
-- Description: Adds schema for the MVP rigorous assessment, retake ledger, and B2B metrics.

-- 1. Assessment Ledgers (The Retake Ledger)
-- Tracks every attempt a candidate makes at the premium vetting gate.
CREATE TABLE IF NOT EXISTS assessment_ledgers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL,
    status VARCHAR(50) NOT NULL, -- e.g., 'passed', 'failed', 'in_progress'
    overall_score INT DEFAULT 0,
    
    -- Technical Evaluation Metrics
    architectural_velocity_score INT DEFAULT 0,
    code_auditing_score INT DEFAULT 0,
    debugging_loop_efficiency_score INT DEFAULT 0,
    
    -- Operational Evaluation Metrics
    sla_punctuality_score INT DEFAULT 0,
    tooling_fluency_score INT DEFAULT 0,
    communication_clarity_score INT DEFAULT 0,
    
    -- Feedback & Growth
    identified_gaps JSONB DEFAULT '[]', -- specific gaps (e.g., "Over-reliance on AI code generation")
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, attempt_number)
);

-- 2. Extend Profiles Table
-- Add specific columns to track the current state of the candidate's employability and vetting status.
ALTER TABLE profiles 
ADD COLUMN IF NOT EXISTS employability_score INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS highest_assessment_score INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS assessment_attempts INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS premium_vetting_passed BOOLEAN DEFAULT FALSE;

-- 3. Assessment Sessions (For the AI-Leveraged Technical Interview)
-- Tracks the specific 60-minute simulated pair-programming sessions.
CREATE TABLE IF NOT EXISTS assessment_sessions (
    id UUID PRIMARY KEY,
    ledger_id UUID NOT NULL REFERENCES assessment_ledgers(id) ON DELETE CASCADE,
    session_type VARCHAR(50) NOT NULL, -- e.g., 'ai_technical', 'hr_behavioral'
    problem_statement TEXT,
    
    -- Raw Metrics collected during the session
    time_to_first_solution_seconds INT,
    bugs_introduced_by_ai INT DEFAULT 0,
    bugs_caught_by_candidate INT DEFAULT 0,
    compilation_attempts INT DEFAULT 0,
    
    -- External AI Service Reference (URL/ID of the analysis)
    ai_analysis_reference_id VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_assessment_ledgers_user_id ON assessment_ledgers(user_id);
CREATE INDEX IF NOT EXISTS idx_assessment_sessions_ledger_id ON assessment_sessions(ledger_id);
CREATE INDEX IF NOT EXISTS idx_profiles_employability ON profiles(employability_score DESC);
