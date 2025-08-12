package main

import (
	"log"
	"os"

	"github.com/Mungge/Fleecy-Cloud/handlers"
	authHandlers "github.com/Mungge/Fleecy-Cloud/handlers/auth"
	"github.com/Mungge/Fleecy-Cloud/initialization"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 애플리케이션 초기화
	err := initialization.InitializeApplication()
	if err != nil {
		log.Fatalf("애플리케이션 초기화 실패: %v", err)
	}

	// 리포지토리 초기화
	repos := initialization.InitializeRepositories()

	// Aggregator 의존성 초기화
	aggregatorDeps := initialization.NewDependencies()

	// 핸들러 초기화
	authHandler := authHandlers.NewAuthHandler(
		repos.UserRepo,
		repos.RefreshTokenRepo,
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
	)
	cloudHandler := handlers.NewCloudHandler(repos.CloudRepo)
	flHandler := handlers.NewFederatedLearningHandler(repos.FLRepo)
	participantHandler := handlers.NewParticipantHandler(repos.ParticipantRepo)
	aggregatorHandler := aggregatorDeps.AggregatorHandler

	// Gin 라우터 설정
	r := gin.Default()

	// CORS 설정
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 라우트 설정 - 각 도메인별로 개별 설정
	// 인증 라우트 (인증 미들웨어 없음)
	routes.SetupAuthRoutes(r, authHandler)

	// 인증이 필요한 라우트 그룹
	authorized := r.Group("/api")
	authorized.Use(middlewares.AuthMiddleware())

	// 각 도메인별 라우트 설정
	routes.SetupCloudRoutes(authorized, cloudHandler)
	routes.SetupParticipantRoutes(authorized, participantHandler)
	routes.SetupFederatedLearningRoutes(authorized, flHandler)
	routes.SetupAggregatorRoutes(authorized, aggregatorHandler)
	
	// VM 라우트 설정 (전체 엔진에 설정, 인증은 내부에서 처리)
	routes.SetupVirtualMachineRoutes(r, repos.VMRepo, repos.ParticipantRepo)

	// 서버 시작
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}