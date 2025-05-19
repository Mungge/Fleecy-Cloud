package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	githubClientID string
	githubClientSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, githubClientID, githubClientSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		githubClientID: githubClientID,
		githubClientSecret: githubClientSecret,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// 비밀번호 유효성 검사
func validatePassword(password string) error {
	var (
		hasMinLen  = len(password) >= 8
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return fmt.Errorf("비밀번호는 최소 8자 이상이어야 합니다")
	}
	if !hasUpper {
		return fmt.Errorf("비밀번호는 대문자를 포함해야 합니다")
	}
	if !hasLower {
		return fmt.Errorf("비밀번호는 소문자를 포함해야 합니다")
	}
	if !hasNumber {
		return fmt.Errorf("비밀번호는 숫자를 포함해야 합니다")
	}
	if !hasSpecial {
		return fmt.Errorf("비밀번호는 특수문자를 포함해야 합니다")
	}

	return nil
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 비밀번호 유효성 검사
	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 이메일 중복 확인
	exists, err := h.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "이메일 확인 중 오류가 발생했습니다"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "이미 사용 중인 이메일입니다"})
		return
	}

	// 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "비밀번호 처리 중 오류가 발생했습니다"})
		return
	}

	// 사용자 생성
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 생성 중 오류가 발생했습니다"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "회원가입이 완료되었습니다"})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 사용자 조회
	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "로그인 처리 중 오류가 발생했습니다"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// 비밀번호 확인
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// JWT 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // TODO: 환경 변수로 이동
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 중 오류가 발생했습니다"})
		return
	}

	response := AuthResponse{
		Token: tokenString,
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

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GitHubLoginHandler(c *gin.Context) {
	// GitHub OAuth URL로 리다이렉트
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
		h.githubClientID,
		"http://localhost:8080/api/auth/github/callback",
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *AuthHandler) GitHubCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "인증 코드가 없습니다"})
		return
	}

	// GitHub 액세스 토큰 요청
	tokenURL := "https://github.com/login/oauth/access_token"
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 요청 생성 실패"})
		return
	}

	q := req.URL.Query()
	q.Add("client_id", h.githubClientID)
	q.Add("client_secret", h.githubClientSecret)
	q.Add("code", code)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub 토큰 요청 실패"})
		return
	}
	defer resp.Body.Close()

	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 응답 파싱 실패"})
		return
	}

	// GitHub 사용자 정보 요청
	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 정보 요청 생성 실패"})
		return
	}

	userReq.Header.Add("Authorization", fmt.Sprintf("token %s", tokenResp.AccessToken))
	userResp, err := client.Do(userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub 사용자 정보 요청 실패"})
		return
	}
	defer userResp.Body.Close()

	var githubUser GitHubUser
	if err := json.NewDecoder(userResp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 정보 파싱 실패"})
		return
	}

	// 이메일이 없는 경우 이메일 정보 요청
	if githubUser.Email == "" {
		emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "이메일 정보 요청 생성 실패"})
			return
		}

		emailReq.Header.Add("Authorization", fmt.Sprintf("token %s", tokenResp.AccessToken))
		emailResp, err := client.Do(emailReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub 이메일 정보 요청 실패"})
			return
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "이메일 정보 파싱 실패"})
			return
		}

		for _, email := range emails {
			if email.Primary && email.Verified {
				githubUser.Email = email.Email
				break
			}
		}
	}

	// 사용자 조회 또는 생성
	user, err := h.userRepo.GetUserByEmail(githubUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 조회 실패"})
		return
	}

	if user == nil {
		// 새 사용자 생성
		user = &models.User{
			Name:  githubUser.Name,
			Email: githubUser.Email,
			// GitHub 로그인 사용자는 비밀번호가 필요 없음
			PasswordHash: "",
		}
		if err := h.userRepo.CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 생성 실패"})
			return
		}
	}

	// JWT 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // TODO: 환경 변수로 이동
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패"})
		return
	}

	response := AuthResponse{
		Token: tokenString,
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

	c.JSON(http.StatusOK, response)
}