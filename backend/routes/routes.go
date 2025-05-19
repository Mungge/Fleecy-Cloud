package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, cloudHandler *handlers.CloudHandler) {
	// 인증 라우트
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.RegisterHandler)
		auth.POST("/login", authHandler.LoginHandler)
		auth.GET("/github", authHandler.GitHubLoginHandler)
		auth.GET("/github/callback", authHandler.GitHubCallbackHandler)
	}

	// 클라우드 라우트
	clouds := r.Group("/api/clouds")
	{
		clouds.GET("", cloudHandler.GetClouds)
		clouds.POST("", cloudHandler.AddCloud)
		clouds.DELETE("/:id", cloudHandler.DeleteCloud)
		clouds.POST("/upload", cloudHandler.UploadCloudCredential)
	}

	// 집계자 라우트 그룹
	aggregator := r.Group("/aggregator")
	{
		aggregator.POST("/estimate", handlers.EstimateHandler)
		aggregator.POST("/recommend", handlers.RecommendHandler)
	}

}