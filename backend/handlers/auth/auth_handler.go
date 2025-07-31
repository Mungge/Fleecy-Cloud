package handlers

import (
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
)

// AuthHandler는 모든 인증 방식을 통합하는 인터페이스입니다
type AuthHandler struct {
	Local  *LocalAuthHandler
	GitHub *GitHubAuthHandler
	Base   *BaseAuthHandler
}

// NewAuthHandler는 새로운 통합 AuthHandler 인스턴스를 생성합니다
func NewAuthHandler(
	userRepo *repository.UserRepository, 
	refreshTokenRepo *repository.RefreshTokenRepository, 
	githubClientID, githubClientSecret string,
) *AuthHandler {
	// 각 핸들러 초기화
	localHandler := NewLocalAuthHandler(userRepo, refreshTokenRepo)
	githubHandler := NewGitHubAuthHandler(userRepo, refreshTokenRepo, githubClientID, githubClientSecret)
	baseHandler := NewBaseAuthHandler(userRepo, refreshTokenRepo)

	return &AuthHandler{
		Local:  localHandler,
		GitHub: githubHandler,
		Base:   baseHandler,
	}
}

// 편의 메서드들 - 기존 라우팅 코드와의 호환성을 위해 제공

// 일반 인증 관련
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	h.Local.RegisterHandler(c)
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	h.Local.LoginHandler(c)
}

func (h *AuthHandler) ChangePasswordHandler(c *gin.Context) {
	h.Local.ChangePasswordHandler(c)
}

func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
	h.Local.ResetPasswordHandler(c)
}

// GitHub OAuth 관련
func (h *AuthHandler) GitHubLoginHandler(c *gin.Context) {
	h.GitHub.GitHubLoginHandler(c)
}

func (h *AuthHandler) GitHubCallbackHandler(c *gin.Context) {
	h.GitHub.GitHubCallbackHandler(c)
}

// 공통 기능 관련
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	h.Base.RefreshTokenHandler(c)
}

func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	h.Base.LogoutHandler(c)
}

func (h *AuthHandler) LogoutAllHandler(c *gin.Context) {
	h.Base.LogoutAllHandler(c)
}

func (h *AuthHandler) GetActiveTokensHandler(c *gin.Context) {
	h.Base.GetActiveTokensHandler(c)
}