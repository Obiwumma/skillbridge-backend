// Package recruiter handles matchmaking calculations between candidate capabilities and enterprise job requirements.
package recruiter

import (
	"context"
	"fmt"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/internal/profiles"
	"skillbridge-backend/pkg/errors"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// RecruiterService coordinates job board listings and candidate capability matching rules.
type RecruiterService interface {
	PostJob(ctx context.Context, title, company, description string, skillsRequired []string) (*JobPosting, error)
	GetJobs(ctx context.Context) ([]JobPosting, error)
	MatchCandidates(ctx context.Context, jobID uuid.UUID) ([]CandidateMatch, error)
}

type recruiterService struct {
	repo           RecruiterRepository
	profileService profiles.ProfileService
}

// NewRecruiterService constructs an implementation of RecruiterService.
func NewRecruiterService(repo RecruiterRepository, profileService profiles.ProfileService) RecruiterService {
	return &recruiterService{
		repo:           repo,
		profileService: profileService,
	}
}

// PostJob registers an enterprise career posting and emits matchmaking alerts to the event network.
func (s *recruiterService) PostJob(ctx context.Context, title, company, description string, skillsRequired []string) (*JobPosting, error) {
	job := &JobPosting{
		ID:             uuid.New(),
		Title:          title,
		Company:        company,
		Description:    description,
		SkillsRequired: skillsRequired,
	}

	err := s.repo.CreateJob(ctx, job)
	if err != nil {
		return nil, errors.NewInternalError("failed to create job posting", err)
	}

	_ = events.Publish(events.JobMatchUpdated, map[string]string{
		"job_id": job.ID.String(),
		"title":  job.Title,
	})

	return job, nil
}

// GetJobs queries all active career posts currently registered.
func (s *recruiterService) GetJobs(ctx context.Context) ([]JobPosting, error) {
	jobs, err := s.repo.GetJobs(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to list jobs", err)
	}
	return jobs, nil
}

// MatchCandidates queries the candidate pool and calculates score metrics relative to job prerequisites.
func (s *recruiterService) MatchCandidates(ctx context.Context, jobID uuid.UUID) ([]CandidateMatch, error) {
	job, err := s.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve job profile", err)
	}
	if job == nil {
		return nil, errors.NewNotFoundError("job posting not found", nil)
	}

	candidates, err := s.repo.GetCandidatesForMatch(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve candidate pool", err)
	}

	var results []CandidateMatch

	for _, c := range candidates {
		profile, err := s.profileService.GetProfile(ctx, c.UserID)
		if err != nil {
			continue
		}

		score, reasons := s.calculateMatchScore(job.SkillsRequired, profile.Skills)

		if c.PremiumVettingPassed {
			reasons = append([]string{fmt.Sprintf("VETTING VERIFIED: Architectural Velocity (%d), Debugging Efficiency (%d), Communication (%d)", 
				c.ArchitecturalVelocity, c.DebuggingEfficiency, c.CommunicationClarity)}, reasons...)
			score += 10 // Bonus for passing rigorous assessment
			if score > 100 {
				score = 100
			}
		}

		c.MatchScore = score
		c.MatchReasons = reasons

		results = append(results, c)
	}

	sort.Slice(results, func(i, j int) bool {
		// Premium vetting trumps standard matching
		if results[i].PremiumVettingPassed != results[j].PremiumVettingPassed {
			return results[i].PremiumVettingPassed
		}
		if results[i].MatchScore == results[j].MatchScore {
			return results[i].EmployabilityScore > results[j].EmployabilityScore
		}
		return results[i].MatchScore > results[j].MatchScore
	})

	return results, nil
}

func (s *recruiterService) calculateMatchScore(requiredSkills []string, userSkills []profiles.SkillItem) (int, []string) {
	if len(requiredSkills) == 0 {
		return 100, []string{"No specific skills required for this position."}
	}

	var scoreSum float64
	var reasons []string

	skillMap := make(map[string]profiles.SkillItem)
	for _, sk := range userSkills {
		skillMap[strings.ToLower(sk.Name)] = sk
	}

	for _, req := range requiredSkills {
		reqLower := strings.ToLower(req)
		if userSk, found := skillMap[reqLower]; found {
			weight := 1.0
			switch userSk.Level {
			case 1:
				weight = 0.4
				reasons = append(reasons, fmt.Sprintf("Has basic knowledge in %s (Level 1)", req))
			case 2:
				weight = 0.75
				reasons = append(reasons, fmt.Sprintf("Intermediate familiarity in %s (Level 2)", req))
			default:
				weight = 1.0
				reasons = append(reasons, fmt.Sprintf("Highly qualified in %s (Level %d)", req, userSk.Level))
			}
			scoreSum += weight
		} else {
			reasons = append(reasons, fmt.Sprintf("Missing required skill: %s", req))
		}
	}

	finalScore := int((scoreSum / float64(len(requiredSkills))) * 100)
	if finalScore == 100 {
		reasons = append(reasons, "Perfect skill match match!")
	}

	return finalScore, reasons
}
