// Package profiles implements student profile structures, skill states, and XP progression logic.
package profiles

import (
	"context"
	"encoding/json"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"

	"github.com/google/uuid"
)

// ProfileService specifies operations managing student profiles, XP tracking, and NATS event synchronizations.
type ProfileService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, university, currentLevel string) (*Profile, error)
	AddXP(ctx context.Context, userID uuid.UUID, skillName string, xpAmount int) (*Profile, error)
	SubscribeToEvents()
}

type profileService struct {
	repo ProfileRepository
}

// NewProfileService aggregates a profileService linked to repositories.
func NewProfileService(repo ProfileRepository) ProfileService {
	return &profileService{repo: repo}
}

// GetProfile returns the student profile matching the user ID, throwing a NotFoundError if missing.
func (s *profileService) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve profile", err)
	}
	if p == nil {
		return nil, errors.NewNotFoundError("profile not found", nil)
	}
	return p, nil
}

// UpdateProfile mutates university names and current education level attributes.
func (s *profileService) UpdateProfile(ctx context.Context, userID uuid.UUID, university, currentLevel string) (*Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve profile", err)
	}
	if p == nil {
		return nil, errors.NewNotFoundError("profile not found", nil)
	}

	p.University = university
	p.CurrentLevel = currentLevel

	err = s.repo.Update(ctx, p)
	if err != nil {
		return nil, errors.NewInternalError("failed to update profile", err)
	}

	return p, nil
}

// AddXP awards XP to a skill, executes level-up routines, and re-calculates the global EmployabilityScore.
func (s *profileService) AddXP(ctx context.Context, userID uuid.UUID, skillName string, xpAmount int) (*Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve profile", err)
	}
	if p == nil {
		return nil, errors.NewNotFoundError("profile not found", nil)
	}

	p.TotalXP += xpAmount

	found := false
	for i, skill := range p.Skills {
		if skill.Name == skillName {
			p.Skills[i].XP += xpAmount
			newLevel := (p.Skills[i].XP / 100) + 1
			if newLevel > p.Skills[i].Level {
				p.Skills[i].Level = newLevel
			}
			found = true
			break
		}
	}

	if !found {
		p.Skills = append(p.Skills, SkillItem{
			Name:  skillName,
			Level: (xpAmount / 100) + 1,
			XP:    xpAmount,
		})
	}

	var skillLevelSum int
	for _, skill := range p.Skills {
		skillLevelSum += skill.Level
	}
	p.EmployabilityScore = (p.TotalXP / 10) + (skillLevelSum * 5)

	err = s.repo.Update(ctx, p)
	if err != nil {
		return nil, errors.NewInternalError("failed to record progression updates", err)
	}

	return p, nil
}

// SubscribeToEvents registers background subscribers to handle registrations and CV analyses automatically.
func (s *profileService) SubscribeToEvents() {
	_, err := events.Subscribe(events.UserRegistered, func(data []byte) {
		logger.Info("Profiles module received user.registered event")

		var payload struct {
			UserID       string `json:"user_id"`
			Role         string `json:"role"`
			University   string `json:"university"`
			CurrentLevel string `json:"current_level"`
		}

		err := json.Unmarshal(data, &payload)
		if err != nil {
			logger.Error("Failed to parse user.registered payload", err)
			return
		}

		if payload.Role != "student" {
			logger.Info("User role is not student, skipping profile creation", "role", payload.Role)
			return
		}

		uid, err := uuid.Parse(payload.UserID)
		if err != nil {
			logger.Error("Invalid user uuid in registration payload", err)
			return
		}

		profile := &Profile{
			UserID:             uid,
			University:         payload.University,
			CurrentLevel:       payload.CurrentLevel,
			EmployabilityScore: 0,
			TotalXP:            0,
			Skills:             []SkillItem{},
		}

		ctx := context.Background()
		err = s.repo.Create(ctx, profile)
		if err != nil {
			logger.Error("Failed to initialize student profile after registration", err)
		} else {
			logger.Info("Successfully initialized student profile", "user_id", payload.UserID)
		}
	})

	if err != nil {
		logger.Error("Profiles module failed to subscribe to user.registered", err)
	}

	_, err = events.Subscribe(events.SkillAnalysisCompleted, func(data []byte) {
		logger.Info("Profiles module received skill.analysis.completed event")

		var payload struct {
			UserID string `json:"user_id"`
			Skills []struct {
				Name  string `json:"name"`
				Level int    `json:"level"`
			} `json:"skills"`
		}

		err := json.Unmarshal(data, &payload)
		if err != nil {
			logger.Error("Failed to parse skill.analysis.completed payload", err)
			return
		}

		uid, err := uuid.Parse(payload.UserID)
		if err != nil {
			logger.Error("Invalid user uuid in skill analysis payload", err)
			return
		}

		ctx := context.Background()
		p, err := s.repo.GetByUserID(ctx, uid)
		if err != nil {
			logger.Error("Failed to retrieve profile to merge skills", err)
			return
		}
		if p == nil {
			logger.Warn("Profile not found for skill merge", "user_id", payload.UserID)
			return
		}

		for _, s := range payload.Skills {
			found := false
			for i, existing := range p.Skills {
				if existing.Name == s.Name {
					if s.Level > existing.Level {
						p.Skills[i].Level = s.Level
					}
					found = true
					break
				}
			}
			if !found {
				p.Skills = append(p.Skills, SkillItem{
					Name:  s.Name,
					Level: s.Level,
					XP:    s.Level * 100,
				})
			}
		}

		var totalXP int
		var skillLevelSum int
		for _, skill := range p.Skills {
			totalXP += skill.XP
			skillLevelSum += skill.Level
		}
		p.TotalXP = totalXP
		p.EmployabilityScore = (p.TotalXP / 10) + (skillLevelSum * 5)

		err = s.repo.Update(ctx, p)
		if err != nil {
			logger.Error("Failed to save merged skills profile", err)
		} else {
			logger.Info("Merged skills into student profile successfully", "user_id", payload.UserID)
		}
	})

	if err != nil {
		logger.Error("Profiles module failed to subscribe to skill.analysis.completed", err)
	}
}
