package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT 토큰에서 사용자 ID를 추출하는 함수
func GetUserIDFromContext(c *gin.Context) (int64, error) {
	// Authorization 헤더 확인
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 토큰 서명 검증 (secret key는 로그인 시 사용한 것과 동일해야 함)
			return []byte("your-secret-key"), nil
		})
		
		if err != nil {
			return 0, fmt.Errorf("토큰 파싱 실패: %v", err)
		}
		
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userIDFloat, ok := claims["user_id"]; ok {
				switch v := userIDFloat.(type) {
				case float64:
					return int64(v), nil
				case json.Number:
					userID, _ := v.Int64()
					return userID, nil
				}
			}
		}
	}
	
	// 2. 쿠키에서 토큰 확인
	tokenCookie, err := c.Cookie("token")
	if err == nil {
		token, err := jwt.Parse(tokenCookie, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})
		
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userIDFloat, ok := claims["user_id"]; ok {
					switch v := userIDFloat.(type) {
					case float64:
						return int64(v), nil
					case json.Number:
						userID, _ := v.Int64()
						return userID, nil
					}
				}
			}
		}
	}
	
	return 0, fmt.Errorf("인증되지 않은 사용자")
} 

func GetUserIDFromMiddleware(c *gin.Context) int64 {
    userIDInterface, exists := c.Get("userID")
    if !exists {
        panic("사용자 ID가 Context에 설정되지 않았습니다")
    }
    return userIDInterface.(int64)
}