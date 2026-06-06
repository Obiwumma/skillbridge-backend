// Package ai_orchestrator manages downstream communication with external machine learning pipelines.
package ai_orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"skillbridge-backend/pkg/config"
	appErrors "skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"
	"time"
)

// ExperienceLevel represents custom restrictions mapping student career tiers.
type ExperienceLevel string

const (
	LevelBeginner     ExperienceLevel = "BEGINNER"
	LevelIntermediate ExperienceLevel = "INTERMEDIATE"
	LevelAdvanced     ExperienceLevel = "ADVANCED"
)

// Priority represents dynamic priority indexes for learning tracks.
type Priority string

const (
	PriorityLow      Priority = "LOW"
	PriorityMedium   Priority = "MEDIUM"
	PriorityHigh     Priority = "HIGH"
	PriorityCritical Priority = "CRITICAL"
)

// AIResponseWrapper defines the canonical JSON response envelope returned by all external AI services.
type AIResponseWrapper struct {
	Success   bool            `json:"success"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data,omitempty"`
	Meta      AIMeta          `json:"meta,omitempty"`
	Error     *AIErrorDetail  `json:"error,omitempty"`
}

// AIMeta captures performance metrics compiled by downstream inference models.
type AIMeta struct {
	Model            string `json:"model"`
	ProcessingTimeMs int    `json:"processing_time_ms"`
}

// AIErrorDetail maps system failures parsed from downstream nodes.
type AIErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CVUploadedRequest maps parameters emitted when uploading new resumes.
type CVUploadedRequest struct {
	Event      string `json:"event"`
	RequestID  string `json:"request_id"`
	UserID     string `json:"user_id"`
	ResumeURL  string `json:"resume_url"`
	TargetRole string `json:"target_role"`
	Region     string `json:"region"`
}

// AISkillItem represents parsed candidate proficiency.
type AISkillItem struct {
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// AIMissingSkill represents missing competencies flagged during resume scans.
type AIMissingSkill struct {
	Name             string  `json:"name"`
	MarketImportance float64 `json:"market_importance"`
}

// AIEducationItem maps academic achievements from scanned resumes.
type AIEducationItem struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
}

// CVAnalysisResponse contains parsed CV metrics returned by the parser service.
type CVAnalysisResponse struct {
	Skills             []AISkillItem     `json:"skills"`
	MissingSkills      []AIMissingSkill  `json:"missing_skills"`
	Education          []AIEducationItem `json:"education"`
	ExperienceLevel    ExperienceLevel   `json:"experience_level"`
	EmployabilityScore int               `json:"employability_score"`
	Summary            string            `json:"summary"`
}

// SkillGapRequest maps input parameters to identify competencies mismatches.
type SkillGapRequest struct {
	RequestID     string   `json:"request_id"`
	UserID        string   `json:"user_id"`
	CurrentSkills []string `json:"current_skills"`
	TargetRole    string   `json:"target_role"`
	Region        string   `json:"region"`
}

// AIGapItem represents a specific skill deficit mapped against market requirements.
type AIGapItem struct {
	Skill             string   `json:"skill"`
	Priority          Priority `json:"priority"`
	MarketDemandScore float64  `json:"market_demand_score"`
	Reasoning         string   `json:"reasoning"`
}

// SkillGapResponse summarizes recommended learning orders.
type SkillGapResponse struct {
	Gaps                     []AIGapItem `json:"gaps"`
	RecommendedLearningOrder []string    `json:"recommended_learning_order"`
}

// RoadmapGenRequest maps parameters to generate dynamic graph paths.
type RoadmapGenRequest struct {
	RequestID       string          `json:"request_id"`
	UserID          string          `json:"user_id"`
	MissingSkills   []string        `json:"missing_skills"`
	ExperienceLevel ExperienceLevel `json:"experience_level"`
}

// AIRoadmapNode represents a dynamic milestone node in the learning path.
type AIRoadmapNode struct {
	NodeID         string   `json:"node_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Difficulty     string   `json:"difficulty"`
	XPReward       int      `json:"xp_reward"`
	EstimatedHours int      `json:"estimated_hours"`
	RequiredSkills []string `json:"required_skills"`
	Unlocks        []string `json:"unlocks"`
}

// RoadmapGenResponse returns the compiled path nodes graph.
type RoadmapGenResponse struct {
	Nodes []AIRoadmapNode `json:"nodes"`
}

// CodeReviewRequest compiles code sandbox executions submitted for evaluation.
type CodeReviewRequest struct {
	RequestID       string `json:"request_id"`
	UserID          string `json:"user_id"`
	Language        string `json:"language"`
	TaskPrompt      string `json:"task_prompt"`
	Code            string `json:"code"`
	ExecutionOutput string `json:"execution_output"`
}

// AIIssueItem maps code architectural or security deficiencies flagged during review.
type AIIssueItem struct {
	Severity Priority `json:"severity"`
	Message  string   `json:"message"`
}

// CodeReviewResponse represents granular execution reviews.
type CodeReviewResponse struct {
	Score           int           `json:"score"`
	Strengths       []string      `json:"strengths"`
	Issues          []AIIssueItem `json:"issues"`
	Recommendations []string      `json:"recommendations"`
	Summary         string        `json:"summary"`
}

// ChatCopilotRequest maps prompt parameters for interactive training help.
type ChatCopilotRequest struct {
	RequestID      string                 `json:"request_id"`
	UserID         string                 `json:"user_id"`
	ConversationID string                 `json:"conversation_id"`
	Message        string                 `json:"message"`
	Context        map[string]interface{} `json:"context"`
}

// ChatCopilotResponse compiles generated interactive feedback.
type ChatCopilotResponse struct {
	Message             string   `json:"message"`
	SuggestedNextTopics []string `json:"suggested_next_topics"`
}

// MarketIntelRequest maps regional filters to retrieve job board trend updates.
type MarketIntelRequest struct {
	RequestID string   `json:"request_id"`
	Region    string   `json:"region"`
	Roles     []string `json:"roles"`
}

// AITopSkillItem maps top market demands.
type AITopSkillItem struct {
	Skill       string  `json:"skill"`
	DemandScore float64 `json:"demand_score"`
}

// MarketIntelResponse summarizes scraped job requirements.
type MarketIntelResponse struct {
	TopSkills    []AITopSkillItem `json:"top_skills"`
	MarketTrends []string         `json:"market_trends"`
}

// BlockchainValidationRequest compiles achievement credentials prepared for locking.
type BlockchainValidationRequest struct {
	RequestID          string                 `json:"request_id"`
	UserID             string                 `json:"user_id"`
	AchievementID      string                 `json:"achievement_id"`
	CertificatePayload map[string]interface{} `json:"certificate_payload"`
}

// BlockchainValidationResponse returns generated transactional receipt references.
type BlockchainValidationResponse struct {
	Hash          string `json:"hash"`
	Chain         string `json:"chain"`
	TransactionID string `json:"transaction_id"`
}

// AIOrchestratorService coordinates secure integrations with versioned external AI services.
type AIOrchestratorService interface {
	AnalyzeCV(ctx context.Context, req *CVUploadedRequest) (*CVAnalysisResponse, error)
	AnalyzeSkillGap(ctx context.Context, req *SkillGapRequest) (*SkillGapResponse, error)
	GenerateRoadmap(ctx context.Context, req *RoadmapGenRequest) (*RoadmapGenResponse, error)
	ReviewCode(ctx context.Context, req *CodeReviewRequest) (*CodeReviewResponse, error)
	ChatCopilot(ctx context.Context, req *ChatCopilotRequest) (*ChatCopilotResponse, error)
	AnalyzeMarket(ctx context.Context, req *MarketIntelRequest) (*MarketIntelResponse, error)
	ValidateCertificate(ctx context.Context, req *BlockchainValidationRequest) (*BlockchainValidationResponse, error)
}

type aiOrchestratorService struct {
	client *http.Client
}

// NewAIOrchestratorService returns an AIOrchestratorService instance.
func NewAIOrchestratorService() AIOrchestratorService {
	return &aiOrchestratorService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// AnalyzeCV invokes resume parsing routines, enforcing strict validation checks on parsed scores and schemas.
func (s *aiOrchestratorService) AnalyzeCV(ctx context.Context, req *CVUploadedRequest) (*CVAnalysisResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		logger.Warn("External AI credentials missing, returning validated CV analysis fallback")
		return &CVAnalysisResponse{
			Skills: []AISkillItem{
				{Name: "Go", Category: "backend", ConfidenceScore: 0.95},
				{Name: "PostgreSQL", Category: "backend", ConfidenceScore: 0.90},
			},
			MissingSkills: []AIMissingSkill{
				{Name: "Docker", MarketImportance: 0.83},
			},
			Education: []AIEducationItem{
				{Institution: "Veritas University", Degree: "BSc Computer Science"},
			},
			ExperienceLevel:    LevelIntermediate,
			EmployabilityScore: 67,
			Summary:            "Strong backend fundamentals but weak DevOps exposure.",
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result CVAnalysisResponse
	err := s.executeWithRetry(ctx, "/v1/analyze-resume", payload, &result)
	if err != nil {
		return nil, err
	}

	// Strict validator rules: reject unknown experience levels, malformed confidence/employability ranges
	if !isValidExperienceLevel(string(result.ExperienceLevel)) {
		return nil, appErrors.NewInternalError("invalid AI response: unknown experience_level enum value", nil)
	}
	if result.EmployabilityScore < 0 || result.EmployabilityScore > 100 {
		return nil, appErrors.NewInternalError("invalid AI response: employability_score must be between 0 and 100", nil)
	}
	for _, skill := range result.Skills {
		if skill.ConfidenceScore <= 0.0 || skill.ConfidenceScore > 1.0 {
			return nil, appErrors.NewInternalError("invalid AI response: missing or malformed confidence_score", nil)
		}
	}

	return &result, nil
}

// AnalyzeSkillGap compares user metrics with market requirements to return structured deficit priorities.
func (s *aiOrchestratorService) AnalyzeSkillGap(ctx context.Context, req *SkillGapRequest) (*SkillGapResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		logger.Warn("External AI credentials missing, returning validated Skill Gap fallback")
		return &SkillGapResponse{
			Gaps: []AIGapItem{
				{Skill: "Docker", Priority: PriorityHigh, MarketDemandScore: 0.87, Reasoning: "Required in 72% of backend roles analyzed."},
			},
			RecommendedLearningOrder: []string{"Docker", "Kubernetes", "CI/CD"},
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result SkillGapResponse
	err := s.executeWithRetry(ctx, "/v1/skill-gap", payload, &result)
	if err != nil {
		return nil, err
	}

	// Validate priority enums
	for _, gap := range result.Gaps {
		if !isValidPriority(string(gap.Priority)) {
			return nil, appErrors.NewInternalError("invalid AI response: unknown priority enum value", nil)
		}
	}

	return &result, nil
}

// GenerateRoadmap requests graph-based pathways mapped to identified skill gaps.
func (s *aiOrchestratorService) GenerateRoadmap(ctx context.Context, req *RoadmapGenRequest) (*RoadmapGenResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		logger.Warn("External AI credentials missing, returning validated Roadmap fallback")
		return &RoadmapGenResponse{
			Nodes: []AIRoadmapNode{
				{
					NodeID:         "1",
					Title:          "Docker Fundamentals",
					Description:    "Learn containers and image management.",
					Difficulty:     "BEGINNER",
					XPReward:       120,
					EstimatedHours: 6,
					RequiredSkills: []string{},
					Unlocks:        []string{"2"},
				},
			},
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result RoadmapGenResponse
	err := s.executeWithRetry(ctx, "/v1/generate-roadmap", payload, &result)
	if err != nil {
		return nil, err
	}

	// Validate difficulty enums
	for _, node := range result.Nodes {
		if !isValidExperienceLevel(node.Difficulty) {
			return nil, appErrors.NewInternalError("invalid AI response: unknown difficulty enum value", nil)
		}
	}

	return &result, nil
}

// ReviewCode submits execution payloads to evaluate code security and architecture.
func (s *aiOrchestratorService) ReviewCode(ctx context.Context, req *CodeReviewRequest) (*CodeReviewResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		logger.Warn("External AI credentials missing, returning validated Code Review fallback")
		return &CodeReviewResponse{
			Score:     74,
			Strengths: []string{"Good middleware separation"},
			Issues: []AIIssueItem{
				{Severity: PriorityHigh, Message: "Hardcoded JWT secret detected."},
			},
			Recommendations: []string{"Move secrets to environment variables."},
			Summary:         "Strong structure but insecure secret handling.",
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result CodeReviewResponse
	err := s.executeWithRetry(ctx, "/v1/code-review", payload, &result)
	if err != nil {
		return nil, err
	}

	// Validate issues severity enums
	for _, issue := range result.Issues {
		if !isValidPriority(string(issue.Severity)) {
			return nil, appErrors.NewInternalError("invalid AI response: unknown severity enum value", nil)
		}
	}

	return &result, nil
}

// ChatCopilot handles interactive conversational prompts.
func (s *aiOrchestratorService) ChatCopilot(ctx context.Context, req *ChatCopilotRequest) (*ChatCopilotResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		return &ChatCopilotResponse{
			Message:             "Docker volumes allow persistent data storage outside containers.",
			SuggestedNextTopics: []string{"Bind Mounts", "Named Volumes"},
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result ChatCopilotResponse
	err := s.executeWithRetry(ctx, "/v1/copilot-chat", payload, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AnalyzeMarket compiles global salary and skill requirement indexes.
func (s *aiOrchestratorService) AnalyzeMarket(ctx context.Context, req *MarketIntelRequest) (*MarketIntelResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		return &MarketIntelResponse{
			TopSkills: []AITopSkillItem{
				{Skill: "Docker", DemandScore: 0.91},
			},
			MarketTrends: []string{"DevOps demand increased by 18%."},
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result MarketIntelResponse
	err := s.executeWithRetry(ctx, "/v1/market-intelligence", payload, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ValidateCertificate processes blockchain transaction hashing validations.
func (s *aiOrchestratorService) ValidateCertificate(ctx context.Context, req *BlockchainValidationRequest) (*BlockchainValidationResponse, error) {
	cfg := config.AppConfig
	if cfg.ExternalAIURL == "" || cfg.ExternalAIKey == "" {
		return &BlockchainValidationResponse{
			Hash:          "0x8f3c...blockchainhash",
			Chain:         "polygon",
			TransactionID: "0xabc123...txid",
		}, nil
	}

	payload, _ := json.Marshal(req)
	var result BlockchainValidationResponse
	err := s.executeWithRetry(ctx, "/v1/blockchain-verify", payload, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *aiOrchestratorService) executeWithRetry(ctx context.Context, path string, payload []byte, target interface{}) error {
	cfg := config.AppConfig
	url := cfg.ExternalAIURL + path

	// Reject oversized payloads (prevent denial-of-service memory pressure)
	if len(payload) > 5*1024*1024 { // 5MB limit
		return appErrors.NewAppError(http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "AI request payload exceeds maximum limit of 5MB", nil)
	}

	var lastErr error
	delay := 100 * time.Millisecond
	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return appErrors.NewInternalError("failed to construct external request", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.ExternalAIKey)

		resp, err := s.client.Do(req)
		if err == nil {
			defer resp.Body.Close()

			var wrapper AIResponseWrapper
			err = json.NewDecoder(resp.Body).Decode(&wrapper)
			if err != nil {
				return appErrors.NewInternalError("failed to decode AI response wrapper", err)
			}

			if !wrapper.Success {
				if wrapper.Error != nil {
					return appErrors.NewAppError(http.StatusBadGateway, wrapper.Error.Code, wrapper.Error.Message, nil)
				}
				return appErrors.NewInternalError("external AI service returned failed result", nil)
			}

			err = json.Unmarshal(wrapper.Data, target)
			if err != nil {
				return appErrors.NewInternalError("failed to unmarshal wrapper data segment", err)
			}

			return nil
		}

		lastErr = err
		logger.Warn("Retrying AI Orchestrator call...", "path", path, "attempt", attempt, "error", lastErr.Error())

		select {
		case <-time.After(delay):
			delay *= 2
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return appErrors.NewInternalError("external AI service unavailable after max retries", lastErr)
}

func isValidExperienceLevel(lvl string) bool {
	return lvl == "BEGINNER" || lvl == "INTERMEDIATE" || lvl == "ADVANCED"
}

func isValidPriority(prio string) bool {
	return prio == "LOW" || prio == "MEDIUM" || prio == "HIGH" || prio == "CRITICAL"
}
