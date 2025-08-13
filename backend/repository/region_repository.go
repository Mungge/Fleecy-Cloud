package repository

import (
	"fmt"

	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RegionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) *RegionRepository {
	return &RegionRepository{db: db}
}

// CreateOrGetRegion은 region이 존재하지 않으면 생성하고, 존재하면 기존 것을 반환합니다
func (r *RegionRepository) CreateOrGetRegion(name string) (*models.Region, error) {
	var region models.Region
	
	// 이름으로 먼저 조회
	err := r.db.Where("name = ?", name).First(&region).Error
	if err == nil {
		return &region, nil
	}
	
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query region: %w", err)
	}
	
	// 존재하지 않으면 생성
	region = models.Region{
		Name: name,
	}
	
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&region).Error; err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}
	
	// 생성 후 다시 조회 (ON CONFLICT DO NOTHING인 경우 ID가 설정되지 않을 수 있음)
	if err := r.db.Where("name = ?", name).First(&region).Error; err != nil {
		return nil, fmt.Errorf("failed to get created region: %w", err)
	}
	
	return &region, nil
}

// GetAllRegions는 모든 region을 조회합니다
func (r *RegionRepository) GetAllRegions() ([]models.Region, error) {
	var regions []models.Region
	if err := r.db.Find(&regions).Error; err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}
	return regions, nil
}

// GetRegionByName은 이름으로 region을 조회합니다
func (r *RegionRepository) GetRegionByName(name string) (*models.Region, error) {
	var region models.Region
	if err := r.db.Where("name = ?", name).First(&region).Error; err != nil {
		return nil, err
	}
	return &region, nil
}
