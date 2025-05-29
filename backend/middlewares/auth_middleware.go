package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/utils"
)

// AuthMiddleware는 JWT 토큰 인증을 확인하는 미들웨어입니다
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// utils에서 사용자 ID 확인
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
} 