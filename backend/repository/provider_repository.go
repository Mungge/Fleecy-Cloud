package repository

import (
	"fmt"

	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// CreateOrGetProvider는 provider가 존재하지 않으면 생성하고, 존재하면 기존 것을 반환합니다
func (r *ProviderRepository) CreateOrGetProvider(name string) (*models.Provider, error) {
	var provider models.Provider
	
	// 이름으로 먼저 조회
	err := r.db.Where("name = ?", name).First(&provider).Error
	if err == nil {
		return &provider, nil
	}
	
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query provider: %w", err)
	}
	
	// 존재하지 않으면 생성
	provider = models.Provider{
		Name: name,
	}
	
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&provider).Error; err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	
	// 생성 후 다시 조회 (ON CONFLICT DO NOTHING인 경우 ID가 설정되지 않을 수 있음)
	if err := r.db.Where("name = ?", name).First(&provider).Error; err != nil {
		return nil, fmt.Errorf("failed to get created provider: %w", err)
	}
	
	return &provider, nil
}

// GetProviderByName은 이름으로 provider를 조회합니다
func (r *ProviderRepository) GetProviderByName(name string) (*models.Provider, error) {
	var provider models.Provider
	if err := r.db.Where("name = ?", name).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetAllProviders는 모든 provider를 조회합니다
func (r *ProviderRepository) GetAllProviders() ([]models.Provider, error) {
	var providers []models.Provider
	if err := r.db.Find(&providers).Error; err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	return providers, nil
}
