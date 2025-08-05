package routes

import (
	authHandlers "github.com/Mungge/Fleecy-Cloud/handlers/auth"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, authHandler *authHandlers.AuthHandler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.RegisterHandler)
		auth.POST("/login", authHandler.LoginHandler)
		auth.POST("/logout", authHandler.LogoutHandler)
		auth.POST("/refresh", authHandler.RefreshTokenHandler)
		auth.GET("/github", authHandler.GitHubLoginHandler)
		auth.GET("/github/callback", authHandler.GitHubCallbackHandler)
	}
}
