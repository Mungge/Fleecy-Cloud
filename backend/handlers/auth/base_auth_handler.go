package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/gin-gonic/gin"
)

// BaseAuthHandler는 모든 인증 핸들러의 공통 기능을 제공합니다
type BaseAuthHandler struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewBaseAuthHandler(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository) *BaseAuthHandler {
	return &BaseAuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// 공통 구조체들
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *BaseAuthHandler) setSecureRefreshTokenCookie(c *gin.Context, refreshToken string, expirationDays int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",           // name
		refreshToken,              // value
		expirationDays*24*3600,   // maxAge (seconds)
		"/api/auth",              // path - 인증 관련 경로로만 제한
		"",                       // domain - 빈 문자열은 현재 도메인 사용
		isProduction(),           // secure - HTTPS에서만 전송 (프로덕션)
		true,                     // httpOnly - JavaScript 접근 불가
	)
}

func (h *BaseAuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",                       // 빈 값
		-1,                       // 즉시 만료
		"/api/auth",              // 동일한 path 사용
		"",                       // 동일한 domain 사용
		isProduction(),
		true,
	)
}

// 토큰 관리 함수들
func (h *BaseAuthHandler) generateAndStoreTokenPair(userID int64, email, name string) (*utils.TokenPair, string, error) {
	tokenPair, refreshTokenString, err := utils.GenerateTokenPair(userID, email, name)
	if err != nil {
		return nil, "", err
	}

	// 리프레시 토큰을 DB에 저장
	refreshTokenExpDays := getRefreshTokenExpirationDays()
	_, err = h.refreshTokenRepo.CreateRefreshToken(userID, refreshTokenString, refreshTokenExpDays)
	if err != nil {
		return nil, "", err
	}

	return tokenPair, refreshTokenString, nil
}

func (h *BaseAuthHandler) createAuthResponse(tokenPair *utils.TokenPair, user *models.User) AuthResponse {
	return AuthResponse{
		AccessToken: tokenPair.AccessToken,
		TokenType:   tokenPair.TokenType,
		ExpiresIn:   tokenPair.ExpiresIn,
		User: struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}
}

// 사용자 관리 함수들
func (h *BaseAuthHandler) findOrCreateUser(email, name string) (*models.User, bool, error) {
	// 기존 사용자 조회
	user, err := h.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, false, err
	}

	if user != nil {
		// 기존 사용자 정보 업데이트 (이름이 다른 경우)
		if user.Name != name && name != "" {
			user.Name = name
			if err := h.userRepo.UpdateUser(user); err != nil {
				log.Printf("사용자 정보 업데이트 실패: %v", err)
			}
		}
		return user, false, nil // 기존 사용자
	}

	// 새 사용자 생성
	newUser := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: "", // OAuth 사용자는 비밀번호 없음
	}

	if err := h.userRepo.CreateUser(newUser); err != nil {
		return nil, false, err
	}

	return newUser, true, nil // 새 사용자
}

// 공통 핸들러들
func (h *BaseAuthHandler) RefreshTokenHandler(c *gin.Context) {
	// 쿠키에서 refresh token 읽기
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "리프레시 토큰이 없습니다"})
		return
	}

	// 리프레시 토큰 검증
	refreshTokenModel, err := h.refreshTokenRepo.GetValidRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 검증 중 오류가 발생했습니다"})
		return
	}

	if refreshTokenModel == nil || !refreshTokenModel.IsValid() {
		log.Printf("유효하지 않은 리프레시 토큰 사용 시도")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 리프레시 토큰입니다"})
		return
	}

	// 새로운 액세스 토큰 생성
	accessToken, err := utils.GenerateAccessToken(refreshTokenModel.UserID, refreshTokenModel.User.Email, refreshTokenModel.User.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "새 토큰 생성 중 오류가 발생했습니다"})
		return
	}

	response := RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(utils.GetAccessTokenExpiration().Seconds()),
	}

	log.Printf("토큰 갱신 성공: 사용자 ID %d", refreshTokenModel.UserID)
	c.JSON(http.StatusOK, response)
}

func (h *BaseAuthHandler) LogoutHandler(c *gin.Context) {
	// 쿠키에서 refresh token 읽기
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		// DB에서 리프레시 토큰 무효화
		if err := h.refreshTokenRepo.RevokeRefreshToken(refreshToken); err != nil {
			log.Printf("리프레시 토큰 무효화 실패: %v", err)
		}
		log.Printf("리프레시 토큰 무효화 완료")
	}

	// 보안 강화된 쿠키 삭제
	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, gin.H{"message": "로그아웃되었습니다"})
}

func (h *BaseAuthHandler) LogoutAllHandler(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	// 사용자의 모든 리프레시 토큰 무효화
	if err := h.refreshTokenRepo.RevokeAllUserRefreshTokens(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "로그아웃 처리 중 오류가 발생했습니다"})
		return
	}

	// 현재 쿠키도 삭제
	h.clearRefreshTokenCookie(c)

	log.Printf("사용자 ID %d 모든 기기에서 로그아웃", userID)
	c.JSON(http.StatusOK, gin.H{"message": "모든 기기에서 로그아웃되었습니다"})
}

func (h *BaseAuthHandler) GetActiveTokensHandler(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	tokens, err := h.refreshTokenRepo.GetUserRefreshTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 목록 조회 중 오류가 발생했습니다"})
		return
	}

	// 민감한 정보 제거 후 반환
	var tokenList []map[string]interface{}
	for _, token := range tokens {
		tokenList = append(tokenList, map[string]interface{}{
			"id":         token.ID,
			"created_at": token.CreatedAt,
			"expires_at": token.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokenList})
}

// 헬퍼 함수들
func isProduction() bool {
	env := os.Getenv("GIN_MODE")
	return env == "release" || env == "production"
}

func getRefreshTokenExpirationDays() int {
	days := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS")
	if days == "" {
		return 7
	}
	
	if d, err := strconv.Atoi(days); err == nil {
		return d
	}
	return 7
}