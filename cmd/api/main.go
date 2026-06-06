// Package main provides the entry point for the Skillbridge Backend API service.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"skillbridge-backend/internal/ai_orchestrator"
	"skillbridge-backend/internal/auth"
	"skillbridge-backend/internal/cache"
	"skillbridge-backend/internal/db"
	"skillbridge-backend/internal/events"
	"skillbridge-backend/internal/middleware"
	"skillbridge-backend/internal/notifications"
	"skillbridge-backend/internal/profiles"
	"skillbridge-backend/internal/recruiter"
	"skillbridge-backend/internal/roadmap"
	"skillbridge-backend/internal/users"
	"skillbridge-backend/internal/workspace"
	"skillbridge-backend/internal/ws"
	"skillbridge-backend/pkg/config"
	"skillbridge-backend/pkg/logger"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	logger.Init()
	config.Load()

	logger.Info("Starting Skillbridge Backend Service in " + config.AppConfig.Env + " mode")

	db.Init()
	cache.Init()
	events.Init()
	ws.Init()

	userRepo := users.NewUserRepository(db.DB)
	userService := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userService)

	authService := auth.NewAuthService(userService)
	authHandler := auth.NewAuthHandler(authService)

	profileRepo := profiles.NewProfileRepository(db.DB)
	profileService := profiles.NewProfileService(profileRepo)
	profileHandler := profiles.NewProfileHandler(profileService)

	roadmapRepo := roadmap.NewRoadmapRepository(db.DB)
	roadmapService := roadmap.NewRoadmapService(roadmapRepo, profileService)
	roadmapHandler := roadmap.NewRoadmapHandler(roadmapService)

	aiService := ai_orchestrator.NewAIOrchestratorService()
	aiHandler := ai_orchestrator.NewAIHandler(aiService)

	workspaceRepo := workspace.NewWorkspaceRepository(db.DB)
	workspaceService := workspace.NewWorkspaceService(workspaceRepo, aiService)
	workspaceHandler := workspace.NewWorkspaceHandler(workspaceService)

	recruiterRepo := recruiter.NewRecruiterRepository(db.DB)
	recruiterService := recruiter.NewRecruiterService(recruiterRepo, profileService)
	recruiterHandler := recruiter.NewRecruiterHandler(recruiterService)

	notificationRepo := notifications.NewNotificationRepository(db.DB)
	notificationService := notifications.NewNotificationService(notificationRepo)

	profileService.SubscribeToEvents()
	roadmapService.SubscribeToEvents()
	aiHandler.SubscribeToEvents()
	notificationService.SubscribeToEvents()

	logger.Info("Decoupled domain NATS event subscribers initialized successfully")

	if config.AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimiter(200, time.Minute))

	router.GET("/ws/realtime", ws.ServeWS)

	api := router.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		private := api.Group("")
		private.Use(middleware.AuthMiddleware(authService))
		{
			private.GET("/user/profile", userHandler.GetProfile)

			private.GET("/profile", profileHandler.GetProfile)
			private.PUT("/profile", profileHandler.UpdateProfile)

			private.GET("/roadmap", roadmapHandler.GetRoadmap)
			private.POST("/roadmap/complete", roadmapHandler.CompleteNode)

			private.POST("/workspace/execute", workspaceHandler.Execute)
			private.GET("/workspace/history", workspaceHandler.GetHistory)

			private.POST("/ai/cv-analyze", aiHandler.CVAnalyze)
			private.POST("/ai/skill-gap", aiHandler.SkillGap)

			recruiterPriv := private.Group("/jobs")
			recruiterPriv.Use(middleware.RequireRole(users.RoleRecruiter, users.RoleAdmin))
			{
				recruiterPriv.POST("", recruiterHandler.PostJob)
				recruiterPriv.GET("/match", recruiterHandler.MatchCandidates)
			}

			private.GET("/jobs", recruiterHandler.GetJobs)
		}
	}

	serverAddr := fmt.Sprintf(":%s", config.AppConfig.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		logger.Info("HTTP server running", "addr", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed to bind", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("API server forced shutdown", err)
	}

	logger.Info("Skillbridge Backend Service terminated cleanly.")
}
