package handlers

import (
	"fmt"
	"log"
	"net/http"
	"unicode"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LocalAuthHandler는 이메일/비밀번호 기반 인증을 처리합니다
type LocalAuthHandler struct {
	*BaseAuthHandler
}

// NewLocalAuthHandler는 새로운 LocalAuthHandler 인스턴스를 생성합니다
func NewLocalAuthHandler(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository) *LocalAuthHandler {
	return &LocalAuthHandler{
		BaseAuthHandler: NewBaseAuthHandler(userRepo, refreshTokenRepo),
	}
}

// 요청 구조체들
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterHandler는 새 사용자 회원가입을 처리합니다
func (h *LocalAuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 비밀번호 유효성 검사
	if err := h.validatePassword(req.Password); err != nil {
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

	log.Printf("새 사용자 회원가입: %s", user.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "회원가입이 완료되었습니다"})
}

// LoginHandler는 이메일/비밀번호 로그인을 처리합니다
func (h *LocalAuthHandler) LoginHandler(c *gin.Context) {
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
		log.Printf("로그인 실패 시도 - 존재하지 않는 이메일: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// GitHub 로그인 사용자인지 확인 (비밀번호가 없는 경우)
	if user.PasswordHash == "" {
		log.Printf("로그인 실패 시도 - OAuth 전용 계정: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "이 계정은 GitHub 로그인을 사용해주세요"})
		return
	}

	// 비밀번호 확인
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("로그인 실패 시도 - 잘못된 비밀번호: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// 토큰 쌍 생성 및 저장
	tokenPair, refreshTokenString, err := h.generateAndStoreTokenPair(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 중 오류가 발생했습니다"})
		return
	}

	// 보안 강화된 쿠키 설정 (RefreshToken만)
	h.setSecureRefreshTokenCookie(c, refreshTokenString, getRefreshTokenExpirationDays())

	// 응답 생성 (AccessToken은 응답 body로만 전달)
	response := h.createAuthResponse(tokenPair, user)

	log.Printf("사용자 로그인 성공: %s", user.Email)
	c.JSON(http.StatusOK, response)
}

// ChangePasswordHandler는 비밀번호 변경을 처리합니다 (추가 기능)
func (h *LocalAuthHandler) ChangePasswordHandler(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 사용자 조회
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "사용자를 찾을 수 없습니다"})
		return
	}

	// GitHub 로그인 사용자인지 확인
	if user.PasswordHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth 계정은 비밀번호 변경이 불가능합니다"})
		return
	}

	// 현재 비밀번호 확인
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "현재 비밀번호가 올바르지 않습니다"})
		return
	}

	// 새 비밀번호 유효성 검사
	if err := h.validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 새 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "비밀번호 처리 중 오류가 발생했습니다"})
		return
	}

	// 비밀번호 업데이트
	user.PasswordHash = string(hashedPassword)
	if err := h.userRepo.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "비밀번호 변경 중 오류가 발생했습니다"})
		return
	}

	// 보안을 위해 모든 기기에서 로그아웃 (선택사항)
	h.refreshTokenRepo.RevokeAllUserRefreshTokens(userID)
	h.clearRefreshTokenCookie(c)

	log.Printf("사용자 비밀번호 변경 성공: %s", user.Email)
	c.JSON(http.StatusOK, gin.H{"message": "비밀번호가 변경되었습니다. 다시 로그인해주세요."})
}

// ResetPasswordHandler는 비밀번호 재설정을 처리합니다 (추가 기능 - 이메일 발송 등은 별도 구현 필요)
func (h *LocalAuthHandler) ResetPasswordHandler(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 사용자 존재 확인
	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "처리 중 오류가 발생했습니다"})
		return
	}

	// 보안상 사용자 존재 여부를 알려주지 않음
	if user == nil || user.PasswordHash == "" {
		// GitHub 계정이거나 존재하지 않는 계정
		c.JSON(http.StatusOK, gin.H{"message": "비밀번호 재설정 이메일을 발송했습니다 (해당 이메일이 존재하는 경우)"})
		return
	}

	// TODO: 실제로는 여기서 비밀번호 재설정 이메일을 발송해야 함
	// 1. 임시 토큰 생성
	// 2. 데이터베이스에 토큰 저장 (만료시간 포함)
	// 3. 이메일 발송 서비스 호출
	
	log.Printf("비밀번호 재설정 요청: %s", req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "비밀번호 재설정 이메일을 발송했습니다 (해당 이메일이 존재하는 경우)"})
}

// 비밀번호 유효성 검사
func (h *LocalAuthHandler) validatePassword(password string) error {
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