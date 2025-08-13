package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
)

type CloudLatencyRepository struct {
	db *gorm.DB
}

func NewCloudLatencyRepository(db *gorm.DB) *CloudLatencyRepository {
	return &CloudLatencyRepository{db: db}
}

func (r *CloudLatencyRepository) GetCloudLatencyByRegion(regionID int) ([]*models.CloudLatency, error) {
	var latencies []*models.CloudLatency
	err := r.db.Preload("SourceProvider").Preload("SourceRegion").Preload("TargetProvider").Preload("TargetRegion").
		Where("source_region_id = ?", regionID).Find(&latencies).Error
	return latencies, err
}

func (r *CloudLatencyRepository) GetCloudLatencyByRegionName(regionName string) ([]*models.CloudLatency, error) {
	var latencies []*models.CloudLatency
	err := r.db.Preload("SourceProvider").Preload("SourceRegion").Preload("TargetProvider").Preload("TargetRegion").
		Joins("JOIN regions ON regions.id = cloud_latency.source_region_id").
		Where("regions.name = ?", regionName).Find(&latencies).Error
	return latencies, err
}

func (r *CloudLatencyRepository) UpdateCloudLatency(latency *models.CloudLatency) error {
	return r.db.Save(latency).Error
}