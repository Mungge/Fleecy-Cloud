package repository

import (
	"gorm.io/gorm"
)

type CloudPriceRepository struct {
	db *gorm.DB
}

func NewCloudPriceRepository(db *gorm.DB) *CloudPriceRepository {
	return &CloudPriceRepository{db: db}
}