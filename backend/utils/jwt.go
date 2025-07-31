package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64 `json:"expires_in"`
}

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	return []byte(secret)
}

// GetAccessTokenExpiration 액세스 토큰 만료 시간 반환 (기본 2시간)
func GetAccessTokenExpiration() time.Duration {
	hours := os.Getenv("ACCESS_TOKEN_EXPIRATION_HOURS")
	if hours == "" {
		return 2 * time.Hour
	}
	
	if h, err := strconv.Atoi(hours); err == nil {
		return time.Duration(h) * time.Hour
	}
	return 2 * time.Hour
}

func GetRefreshTokenExpiration() time.Duration {
	days := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS")
	if days == "" {
		return 7 * 24 * time.Hour
	}
	
	if d, err := strconv.Atoi(days); err == nil {
		return time.Duration(d) * 24 * time.Hour
	}
	return 7 * 24 * time.Hour
}

// GenerateAccessToken 액세스 토큰 생성
func GenerateAccessToken(userID int64, email, name string) (string, error) {
	expirationTime := time.Now().Add(GetAccessTokenExpiration())
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"name":    name,
		"type":    "access",
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	})
	
	return token.SignedString(GetJWTSecret())
}


// GenerateTokenPair 액세스 토큰과 리프레시 토큰 쌍 생성
func GenerateTokenPair(userID int64, email, name string) (*TokenPair, string, error) {
	// 액세스 토큰 생성
	accessToken, err := GenerateAccessToken(userID, email, name)
	if err != nil {
		return nil, "", fmt.Errorf("액세스 토큰 생성 실패: %v", err)
	}
	
	// 리프레시 토큰 생성
	refreshToken := generateRandomToken()
	
	tokenPair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(GetAccessTokenExpiration().Seconds()),
	}
	
	return tokenPair, refreshToken, nil
}

// generateRandomToken 랜덤 토큰 생성
func generateRandomToken() string {
	// 간단한 랜덤 토큰 생성 (실제로는 crypto/rand 사용 권장)
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// ValidateAccessToken 액세스 토큰 검증
func ValidateAccessToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("예상치 못한 서명 방법: %v", token.Header["alg"])
		}
		return GetJWTSecret(), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("토큰 파싱 실패: %v", err)
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 토큰 타입 확인
		if tokenType, exists := claims["type"]; !exists || tokenType != "access" {
			return nil, fmt.Errorf("잘못된 토큰 타입")
		}
		return &claims, nil
	}
	
	return nil, fmt.Errorf("유효하지 않은 토큰")
}

// JWT 토큰에서 사용자 ID를 추출하는 함수
func GetUserIDFromContext(c *gin.Context) (int64, error) {
	// Authorization 헤더 확인
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 토큰 서명 검증 (GetJWTSecret() 사용)
			return GetJWTSecret(), nil
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
			return GetJWTSecret(), nil
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