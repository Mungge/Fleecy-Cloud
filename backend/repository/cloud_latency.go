package repository

import (
	"gorm.io/gorm"
	"github.com/Mungge/Fleecy-Cloud/models"
)

type CloudLatencyRepository struct {
	db *gorm.DB
}

func NewCloudLatencyRepository(db *gorm.DB) *CloudLatencyRepository {
	return &CloudLatencyRepository{db: db}
}

func (r *CloudLatencyRepository) GetCloudLatencyByRegion(region string) ([]*models.CloudLatency, error) {
	var Latency []*models.CloudLatency
	err := r.db.Where("source_region = ?", region).Find(&Latency).Error
	return Latency, err
}

func (r *CloudLatencyRepository) UpdateCloudLatency(latency *models.CloudLatency) error {
	return r.db.Save(latency).Error
}