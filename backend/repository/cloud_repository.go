package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CloudRepository struct {
	db *gorm.DB
}

func NewCloudRepository(db *gorm.DB) *CloudRepository {
	return &CloudRepository{db: db}
}

func (r *CloudRepository) CreateCloudConnection(conn *models.CloudConnection) error {
	// UUID 생성
	conn.ID = uuid.New().String()
	return r.db.Create(conn).Error
}

func (r *CloudRepository) GetCloudConnectionsByUserID(userID int64) ([]*models.CloudConnection, error) {
	var connections []*models.CloudConnection
	err := r.db.Where("user_id = ?", userID).Find(&connections).Error
	return connections, err
}

func (r *CloudRepository) GetCloudConnectionByID(id string) (*models.CloudConnection, error) {
	var conn models.CloudConnection
	err := r.db.Where("id = ?", id).First(&conn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conn, nil
}

func (r *CloudRepository) UpdateCloudConnectionStatus(id string, status string) error {
	return r.db.Model(&models.CloudConnection{}).Where("id = ?", id).Update("status", status).Error
}

func (r *CloudRepository) UpdateCloudConnection(conn *models.CloudConnection) error {
	return r.db.Save(conn).Error
}

func (r *CloudRepository) DeleteCloudConnection(id string) error {
	return r.db.Delete(&models.CloudConnection{}, "id = ?", id).Error
}