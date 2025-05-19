package repository

import (
	"database/sql"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
)

type CloudRepository struct {
	db *sql.DB
}

func NewCloudRepository(db *sql.DB) *CloudRepository {
	return &CloudRepository{db: db}
}

func (r *CloudRepository) CreateCloudConnection(conn *models.CloudConnection) error {
	query := `
		INSERT INTO cloud_connections (
			user_id, provider, name, region, status,
			credential_file, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	return r.db.QueryRow(
		query,
		conn.UserID,
		conn.Provider,
		conn.Name,
		conn.Region,
		conn.Status,
		conn.CredentialFile,
		now,
		now,
	).Scan(&conn.ID)
}

func (r *CloudRepository) GetCloudConnectionsByUserID(userID int64) ([]*models.CloudConnection, error) {
	query := `
		SELECT id, user_id, provider, name, region, status, created_at, updated_at
		FROM cloud_connections
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*models.CloudConnection
	for rows.Next() {
		conn := &models.CloudConnection{}
		err := rows.Scan(
			&conn.ID,
			&conn.UserID,
			&conn.Provider,
			&conn.Name,
			&conn.Region,
			&conn.Status,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *CloudRepository) GetCloudConnectionByID(id int64) (*models.CloudConnection, error) {
	query := `
		SELECT id, user_id, provider, name, region, status,
			   credential_file, created_at, updated_at
		FROM cloud_connections
		WHERE id = $1`

	conn := &models.CloudConnection{}
	err := r.db.QueryRow(query, id).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.Provider,
		&conn.Name,
		&conn.Region,
		&conn.Status,
		&conn.CredentialFile,
		&conn.CreatedAt,
		&conn.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *CloudRepository) UpdateCloudConnectionStatus(id int64, status string) error {
	query := `
		UPDATE cloud_connections
		SET status = $1, updated_at = $2
		WHERE id = $3`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

func (r *CloudRepository) DeleteCloudConnection(id int64) error {
	query := `DELETE FROM cloud_connections WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}