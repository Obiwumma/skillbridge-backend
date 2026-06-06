// Package roadmap implements branching graph architectures representing adaptive student learning pathways.
package roadmap

import (
	"context"
	"encoding/json"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/internal/profiles"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/logger"

	"github.com/google/uuid"
)

// RoadmapService coordinates learning pathway retrievals, node progression states, and NATS event syncs.
type RoadmapService interface {
	GetRoadmap(ctx context.Context, userID uuid.UUID) (*Roadmap, error)
	CompleteNode(ctx context.Context, userID uuid.UUID, nodeID string) (*Roadmap, error)
	SubscribeToEvents()
}

type roadmapService struct {
	repo           RoadmapRepository
	profileService profiles.ProfileService
}

// NewRoadmapService compiles a roadmapService instance coupled with Profile dependencies.
func NewRoadmapService(repo RoadmapRepository, profileService profiles.ProfileService) RoadmapService {
	return &roadmapService{
		repo:           repo,
		profileService: profileService,
	}
}

// GetRoadmap resolves student DAGs, initializing default learning pathways for new accounts.
func (s *roadmapService) GetRoadmap(ctx context.Context, userID uuid.UUID) (*Roadmap, error) {
	rm, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve learning roadmap", err)
	}

	if rm == nil {
		rm = s.generateDefaultRoadmap(userID)
		err = s.repo.Create(ctx, rm)
		if err != nil {
			return nil, errors.NewInternalError("failed to initialize default learning path", err)
		}
		
		_ = events.Publish(events.RoadmapGenerated, map[string]string{
			"user_id":    userID.String(),
			"roadmap_id": rm.ID.String(),
		})
	}

	return rm, nil
}

// CompleteNode marks a dynamic learning track milestone as complete, awarding XP and evaluating child node unlock gates.
func (s *roadmapService) CompleteNode(ctx context.Context, userID uuid.UUID, nodeID string) (*Roadmap, error) {
	rm, err := s.GetRoadmap(ctx, userID)
	if err != nil {
		return nil, err
	}

	node, exists := rm.Nodes[nodeID]
	if !exists {
		return nil, errors.NewNotFoundError("roadmap node not found", nil)
	}

	if node.Status == "completed" {
		return rm, nil
	}

	if node.Status == "locked" {
		return nil, errors.NewValidationError("cannot complete locked node; prerequisites or skill unlock conditions not met", nil)
	}

	node.Status = "completed"

	profile, err := s.profileService.GetProfile(ctx, userID)
	if err == nil {
		for _, skillName := range node.SkillsUnlocked {
			profile, err = s.profileService.AddXP(ctx, userID, skillName, node.XPReward)
			if err != nil {
				logger.Error("Failed to add XP to skill "+skillName, err)
			}
		}
	}

	for _, childID := range node.Children {
		childNode, ok := rm.Nodes[childID]
		if !ok {
			continue
		}

		prereqsMet := true
		for _, preID := range childNode.Prerequisites {
			if rm.Nodes[preID].Status != "completed" {
				prereqsMet = false
				break
			}
		}

		if prereqsMet {
			skillsMet := true
			if profile != nil {
				for _, cond := range childNode.UnlockConditions {
					hasSkill := false
					for _, userSkill := range profile.Skills {
						if userSkill.Name == cond.SkillName {
							if userSkill.Level < cond.MinLevelRequired {
								skillsMet = false
							}
							hasSkill = true
							break
						}
					}
					if !hasSkill {
						skillsMet = false
					}
				}
			}

			if skillsMet {
				childNode.Status = "unlocked"
			}
		}
	}

	err = s.repo.Update(ctx, rm)
	if err != nil {
		return nil, errors.NewInternalError("failed to persist updated learning path", err)
	}

	return rm, nil
}

// SubscribeToEvents registers background listeners to unlock nodes automatically when CV skill analysis yields proficiencies.
func (s *roadmapService) SubscribeToEvents() {
	_, err := events.Subscribe(events.SkillAnalysisCompleted, func(data []byte) {
		logger.Info("Roadmap module received skill.analysis.completed event")

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
		rm, err := s.GetRoadmap(ctx, uid)
		if err != nil {
			logger.Error("Failed to retrieve roadmap for updating skill gaps", err)
			return
		}

		profile, err := s.profileService.GetProfile(ctx, uid)
		if err != nil {
			logger.Error("Failed to fetch profile to unlock roadmap nodes", err)
			return
		}

		updated := false
		for _, node := range rm.Nodes {
			if node.Status != "locked" {
				continue
			}

			prereqsMet := true
			for _, preID := range node.Prerequisites {
				if rm.Nodes[preID].Status != "completed" {
					prereqsMet = false
					break
				}
			}

			if !prereqsMet {
				continue
			}

			skillsMet := true
			for _, cond := range node.UnlockConditions {
				hasSkill := false
				for _, userSkill := range profile.Skills {
					if userSkill.Name == cond.SkillName {
						if userSkill.Level < cond.MinLevelRequired {
							skillsMet = false
						}
						hasSkill = true
						break
					}
				}
				if !hasSkill {
					skillsMet = false
				}
			}

			if skillsMet {
				node.Status = "unlocked"
				updated = true
			}
		}

		if updated {
			err = s.repo.Update(ctx, rm)
			if err != nil {
				logger.Error("Failed to update roadmap states after skill analysis", err)
			} else {
				logger.Info("Roadmap nodes auto-unlocked following skill analysis", "user_id", payload.UserID)
			}
		}
	})

	if err != nil {
		logger.Error("Roadmap module failed to subscribe to skill.analysis.completed", err)
	}
}

func (s *roadmapService) generateDefaultRoadmap(userID uuid.UUID) *Roadmap {
	nodes := map[string]*RoadmapNode{
		"1": {
			ID:            "1",
			Title:         "Go Basics",
			Description:   "Master Go syntax, types, functions, slices, maps, and basic concurrency.",
			Prerequisites: []string{},
			UnlockConditions: []UnlockCondition{},
			XPReward:       100,
			SkillsUnlocked: []string{"Go"},
			Children:       []string{"2", "3"},
			Status:         "unlocked",
		},
		"2": {
			ID:            "2",
			Title:         "Gin HTTP framework & RESTful APIs",
			Description:   "Create HTTP services using Gin, configure routers, capture params, validate JSON payload structure, and use HTTP status codes.",
			Prerequisites: []string{"1"},
			UnlockConditions: []UnlockCondition{
				{SkillName: "Go", MinLevelRequired: 1},
			},
			XPReward:       150,
			SkillsUnlocked: []string{"Go", "REST"},
			Children:       []string{"4"},
			Status:         "locked",
		},
		"3": {
			ID:            "3",
			Title:         "SQL & PostgreSQL Database Integration",
			Description:   "Configure database schema mappings, write clean SQL transactions, perform inner/outer joins, and construct efficient queries.",
			Prerequisites: []string{"1"},
			UnlockConditions: []UnlockCondition{
				{SkillName: "Go", MinLevelRequired: 1},
			},
			XPReward:       150,
			SkillsUnlocked: []string{"SQL", "PostgreSQL"},
			Children:       []string{"4"},
			Status:         "locked",
		},
		"4": {
			ID:            "4",
			Title:         "Advanced Modular Monoliths",
			Description:   "Assemble modular services inside Go, handle caching in Redis, dockerize backends, and handle async workflows using NATS.",
			Prerequisites: []string{"2", "3"},
			UnlockConditions: []UnlockCondition{
				{SkillName: "REST", MinLevelRequired: 1},
				{SkillName: "SQL", MinLevelRequired: 1},
			},
			XPReward:       300,
			SkillsUnlocked: []string{"Docker", "Redis", "NATS"},
			Children:       []string{},
			Status:         "locked",
		},
	}

	return &Roadmap{
		ID:     uuid.New(),
		UserID: userID,
		Title:  "Backend Engineering Core Pathway",
		Nodes:  nodes,
	}
}
