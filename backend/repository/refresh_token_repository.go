// backend/repository/refresh_token_repository.go
package repository

import (
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// CreateRefreshToken 새로운 리프레시 토큰 생성
func (r *RefreshTokenRepository) CreateRefreshToken(userID int64, token string, expirationDays int) (*models.RefreshToken, error) {
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour),
		IsRevoked: false,
	}
	
	err := r.db.Create(refreshToken).Error
	return refreshToken, err
}

// GetValidRefreshToken 유효한 리프레시 토큰 조회
func (r *RefreshTokenRepository) GetValidRefreshToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("token = ? AND is_revoked = ? AND expires_at > ?", 
		token, false, time.Now()).
		Preload("User").
		First(&refreshToken).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &refreshToken, nil
}

// RevokeRefreshToken 리프레시 토큰 무효화
func (r *RefreshTokenRepository) RevokeRefreshToken(token string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"updated_at": time.Now(),
		}).Error
}

// RevokeAllUserRefreshTokens 사용자의 모든 리프레시 토큰 무효화
func (r *RefreshTokenRepository) RevokeAllUserRefreshTokens(userID int64) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"updated_at": time.Now(),
		}).Error
}

// CleanupExpiredTokens 만료된 토큰들 정리
func (r *RefreshTokenRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.RefreshToken{}).Error
}

// GetUserRefreshTokens 사용자의 활성 리프레시 토큰 목록 조회
func (r *RefreshTokenRepository) GetUserRefreshTokens(userID int64) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	err := r.db.Where("user_id = ? AND is_revoked = ? AND expires_at > ?", 
		userID, false, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	
	return tokens, err
}