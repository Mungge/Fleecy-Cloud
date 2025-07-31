package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
)

// GitHubAuthHandler는 GitHub OAuth 인증을 처리합니다
type GitHubAuthHandler struct {
	*BaseAuthHandler
	clientID     string
	clientSecret string
}

// NewGitHubAuthHandler는 새로운 GitHubAuthHandler 인스턴스를 생성합니다
func NewGitHubAuthHandler(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository, clientID, clientSecret string) *GitHubAuthHandler {
	return &GitHubAuthHandler{
		BaseAuthHandler: NewBaseAuthHandler(userRepo, refreshTokenRepo),
		clientID:        clientID,
		clientSecret:    clientSecret,
	}
}

// GitHub API 응답 구조체들
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GitHubLoginHandler는 GitHub OAuth 인증을 시작합니다
func (h *GitHubAuthHandler) GitHubLoginHandler(c *gin.Context) {
	// 상태 값 생성 (CSRF 공격 방어)
	state := h.generateRandomState()
	
	// 상태 값을 임시로 쿠키에 저장
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("oauth_state", state, 600, "/", "", isProduction(), true) // 10분 후 만료
	
	// GitHub OAuth URL로 리다이렉트
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		h.clientID,
		h.getCallbackURL(),
		state,
	)
	
	log.Printf("GitHub OAuth 인증 시작")
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GitHubCallbackHandler는 GitHub OAuth 콜백을 처리합니다
func (h *GitHubAuthHandler) GitHubCallbackHandler(c *gin.Context) {
	// 인증 코드와 상태 값 확인
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "인증 코드가 없습니다"})
		return
	}
	
	// 상태 값 검증 (CSRF 공격 방어)
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		log.Printf("GitHub OAuth CSRF 공격 시도 감지")
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 상태 값입니다"})
		return
	}
	
	// 상태 쿠키 삭제
	c.SetCookie("oauth_state", "", -1, "/", "", isProduction(), true)
	
	// GitHub 액세스 토큰 요청
	githubToken, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 교환 실패: " + err.Error()})
		return
	}
	
	// GitHub 사용자 정보 가져오기
	githubUser, err := h.getGitHubUserInfo(githubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 정보 조회 실패: " + err.Error()})
		return
	}
	
	// 이메일 정보가 없는 경우 별도 요청
	if githubUser.Email == "" {
		email, err := h.getGitHubUserEmail(githubToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "이메일 정보 조회 실패: " + err.Error()})
			return
		}
		githubUser.Email = email
	}
	
	if githubUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub 계정에 공개 이메일이 필요합니다"})
		return
	}
	
	// 사용자 조회 또는 생성
	user, isNewUser, err := h.findOrCreateUser(githubUser.Email, githubUser.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 처리 실패: " + err.Error()})
		return
	}
	
	// 토큰 쌍 생성 및 저장
	tokenPair, refreshTokenString, err := h.generateAndStoreTokenPair(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패: " + err.Error()})
		return
	}
	
	// 보안 강화된 쿠키 설정 (Refresh Token만)
	h.setSecureRefreshTokenCookie(c, refreshTokenString, getRefreshTokenExpirationDays())
	
	// 응답 생성 (Access Token은 응답 body로만 전달)
	response := h.createAuthResponse(tokenPair, user)
	
	if isNewUser {
		log.Printf("GitHub을 통한 새 사용자 생성: %s", user.Email)
	} else {
		log.Printf("GitHub OAuth 로그인 성공: %s", user.Email)
	}
	
	c.JSON(http.StatusOK, response)
}

// exchangeCodeForToken은 GitHub에서 인증 코드를 액세스 토큰으로 교환합니다
func (h *GitHubAuthHandler) exchangeCodeForToken(code string) (string, error) {
	tokenURL := "https://github.com/login/oauth/access_token"
	
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("토큰 요청 생성 실패: %v", err)
	}
	
	// URL 파라미터 설정
	q := req.URL.Query()
	q.Add("client_id", h.clientID)
	q.Add("client_secret", h.clientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", h.getCallbackURL())
	req.URL.RawQuery = q.Encode()
	
	// JSON 응답 요청
	req.Header.Add("Accept", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub 토큰 요청 실패: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 토큰 요청 실패: HTTP %d", resp.StatusCode)
	}
	
	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("토큰 응답 파싱 실패: %v", err)
	}
	
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("액세스 토큰을 받지 못했습니다")
	}
	
	return tokenResp.AccessToken, nil
}

// getGitHubUserInfo는 GitHub 사용자 정보를 가져옵니다
func (h *GitHubAuthHandler) getGitHubUserInfo(token string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("사용자 정보 요청 생성 실패: %v", err)
	}
	
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub 사용자 정보 요청 실패: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 사용자 정보 요청 실패: HTTP %d", resp.StatusCode)
	}
	
	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("사용자 정보 파싱 실패: %v", err)
	}
	
	return &githubUser, nil
}

// getGitHubUserEmail은 GitHub 사용자의 이메일 주소를 가져옵니다
func (h *GitHubAuthHandler) getGitHubUserEmail(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("이메일 정보 요청 생성 실패: %v", err)
	}
	
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub 이메일 정보 요청 실패: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 이메일 정보 요청 실패: HTTP %d", resp.StatusCode)
	}
	
	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("이메일 정보 파싱 실패: %v", err)
	}
	
	// 주 이메일이면서 검증된 이메일 찾기
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	
	// 주 이메일이 없으면 첫 번째 검증된 이메일 사용
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}
	
	return "", fmt.Errorf("검증된 이메일 주소를 찾을 수 없습니다")
}

// 헬퍼 함수들
func (h *GitHubAuthHandler) generateRandomState() string {
	// 간단한 상태 값 생성 (실제로는 더 강력한 랜덤 생성기 사용 권장)
	return fmt.Sprintf("state_%d", time.Now().UnixNano())
}

func (h *GitHubAuthHandler) getCallbackURL() string {
	// 환경변수에서 콜백 URL 가져오기
	callbackURL := os.Getenv("GITHUB_CALLBACK_URL")
	if callbackURL == "" {
		// 기본값 (개발 환경)
		callbackURL = "http://localhost:8080/api/auth/github/callback"
	}
	return callbackURL
}