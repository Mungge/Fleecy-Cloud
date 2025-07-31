package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, authHandler *handlers.AuthHandler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.RegisterHandler)
		auth.POST("/login", authHandler.LoginHandler)
		auth.GET("/github", authHandler.GitHubLoginHandler)
		auth.GET("/github/callback", authHandler.GitHubCallbackHandler)
	}
}
