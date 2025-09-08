package main

import (
	"log"
	"os"
	"time"

	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/Mungge/Fleecy-Cloud/handlers/aggregator"
	authHandlers "github.com/Mungge/Fleecy-Cloud/handlers/auth"
	"github.com/Mungge/Fleecy-Cloud/initialization"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/routes"
	"github.com/Mungge/Fleecy-Cloud/services"

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

	// SSH 키페어 핸들러 초기화
	sshKeypairService := services.NewSSHKeypairService(repos.SSHKeypairRepo)
	sshKeypairHandler := handlers.NewSSHKeypairHandler(sshKeypairService)

	// 핸들러 초기화
	authHandler := authHandlers.NewAuthHandler(
		repos.UserRepo,
		repos.RefreshTokenRepo,
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
	)
	cloudHandler := handlers.NewCloudHandler(repos.CloudRepo)
	flHandler := handlers.NewFederatedLearningHandler(repos.FLRepo, repos.ParticipantRepo, repos.AggregatorRepo, sshKeypairService)
	participantHandler := handlers.NewParticipantHandler(repos.ParticipantRepo)
	aggregatorHandler := aggregatorDeps.AggregatorHandler

	// SSH 키페어 핸들러 초기화
	sshKeypairService = services.NewSSHKeypairService(repos.SSHKeypairRepo)
	sshKeypairHandler = handlers.NewSSHKeypairHandler(sshKeypairService)

	// MLflow 핸들러 초기화
	mlflowURL := os.Getenv("MLFLOW_TRACKING_URI")
	if mlflowURL == "" {
		mlflowURL = "http://localhost:5001" // 기본값
	}
	mlflowHandler := aggregator.NewMLflowHandler(mlflowURL,repos.AggregatorRepo)
	log.Printf("MLflow 서버 URL: %s", mlflowURL)

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

	// 헬스체크 엔드포인트 (인증 불필요)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":     "ok",
			"mlflow_url": mlflowURL,
			"timestamp":  time.Now().Unix(),
			"version":    "1.0.0",
		})
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
	routes.SetupAggregatorRoutes(authorized, aggregatorHandler, mlflowHandler)
	routes.SetupSSHKeypairRoutes(authorized, sshKeypairHandler)

	// VM 라우트 설정 (전체 엔진에 설정, 인증은 내부에서 처리)
	routes.SetupVirtualMachineRoutes(r, repos.ParticipantRepo)

	// 서버 시작 정보 로깅
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("서버 시작...")
	log.Printf("포트: %s", port)
	log.Printf("MLflow URL: %s", mlflowURL)
	log.Printf("환경: %s", gin.Mode())
	log.Printf("CORS 허용 도메인: http://localhost:3000")
	
	// 서버 시작
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
