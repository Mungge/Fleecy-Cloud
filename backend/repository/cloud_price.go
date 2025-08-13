package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
)

type CloudPriceRepository struct {
	db *gorm.DB
}

func NewCloudPriceRepository(db *gorm.DB) *CloudPriceRepository {
	return &CloudPriceRepository{db: db}
}

// GetAllCloudPrices는 모든 클라우드 가격 정보를 조회합니다
func (r *CloudPriceRepository) GetAllCloudPrices() ([]*models.CloudPrice, error) {
	var prices []*models.CloudPrice
	err := r.db.Preload("Provider").Preload("Region").Find(&prices).Error
	return prices, err
}

// GetCloudPricesByProvider는 특정 provider의 가격 정보를 조회합니다
func (r *CloudPriceRepository) GetCloudPricesByProvider(providerID int) ([]*models.CloudPrice, error) {
	var prices []*models.CloudPrice
	err := r.db.Preload("Provider").Preload("Region").
		Where("provider_id = ?", providerID).Find(&prices).Error
	return prices, err
}

// GetCloudPricesByRegion은 특정 region의 가격 정보를 조회합니다
func (r *CloudPriceRepository) GetCloudPricesByRegion(regionID int) ([]*models.CloudPrice, error) {
	var prices []*models.CloudPrice
	err := r.db.Preload("Provider").Preload("Region").
		Where("region_id = ?", regionID).Find(&prices).Error
	return prices, err
}

// GetCloudPricesByProviderAndRegion은 특정 provider와 region의 가격 정보를 조회합니다
func (r *CloudPriceRepository) GetCloudPricesByProviderAndRegion(providerID, regionID int) ([]*models.CloudPrice, error) {
	var prices []*models.CloudPrice
	err := r.db.Preload("Provider").Preload("Region").
		Where("provider_id = ? AND region_id = ?", providerID, regionID).Find(&prices).Error
	return prices, err
}