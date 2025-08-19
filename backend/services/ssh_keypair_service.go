package services

import (
	"fmt"
	"log"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"gorm.io/gorm"
)

type SSHKeypairService struct {
	repo *repository.SSHKeypairRepository
}

func NewSSHKeypairService(repo *repository.SSHKeypairRepository) *SSHKeypairService {
	return &SSHKeypairService{repo: repo}
}

// SaveKeypair SSH 키페어를 암호화하여 저장
func (s *SSHKeypairService) SaveKeypair(aggregatorID, keyName, cloudProvider, region, publicKey, privateKey string) (*models.SSHKeypairResponse, error) {
	// Private Key 암호화
	encryptedPrivateKey, err := utils.EncryptPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %v", err)
	}

	// 기존 키페어 확인
	existingKeypair, err := s.repo.GetKeypairByAggregatorID(aggregatorID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing keypair: %v", err)
	}

	if existingKeypair != nil {
		// 기존 키페어 업데이트
		existingKeypair.KeyName = keyName
		existingKeypair.CloudProvider = cloudProvider
		existingKeypair.Region = region
		existingKeypair.PublicKey = publicKey
		existingKeypair.EncryptedPrivateKey = encryptedPrivateKey

		if err := s.repo.UpdateKeypair(existingKeypair); err != nil {
			return nil, fmt.Errorf("failed to update keypair: %v", err)
		}

		log.Printf("Updated SSH keypair for aggregator: %s", aggregatorID)
		return &models.SSHKeypairResponse{
			ID:            existingKeypair.ID,
			AggregatorID:  existingKeypair.AggregatorID,
			KeyName:       existingKeypair.KeyName,
			CloudProvider: existingKeypair.CloudProvider,
			Region:        existingKeypair.Region,
			PublicKey:     existingKeypair.PublicKey,
			CreatedAt:     existingKeypair.CreatedAt,
			UpdatedAt:     existingKeypair.UpdatedAt,
		}, nil
	}

	// 새 키페어 생성
	newKeypair := &models.SSHKeypair{
		AggregatorID:        aggregatorID,
		KeyName:             keyName,
		CloudProvider:       cloudProvider,
		Region:              region,
		PublicKey:           publicKey,
		EncryptedPrivateKey: encryptedPrivateKey,
	}

	if err := s.repo.CreateKeypair(newKeypair); err != nil {
		return nil, fmt.Errorf("failed to create keypair: %v", err)
	}

	log.Printf("Created SSH keypair for aggregator: %s", aggregatorID)
	return &models.SSHKeypairResponse{
		ID:            newKeypair.ID,
		AggregatorID:  newKeypair.AggregatorID,
		KeyName:       newKeypair.KeyName,
		CloudProvider: newKeypair.CloudProvider,
		Region:        newKeypair.Region,
		PublicKey:     newKeypair.PublicKey,
		CreatedAt:     newKeypair.CreatedAt,
		UpdatedAt:     newKeypair.UpdatedAt,
	}, nil
}

// GetKeypairByAggregatorID 집계자 ID로 키페어 조회 (Private Key 제외)
func (s *SSHKeypairService) GetKeypairByAggregatorID(aggregatorID string) (*models.SSHKeypairResponse, error) {
	keypair, err := s.repo.GetKeypairByAggregatorID(aggregatorID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("keypair not found for aggregator: %s", aggregatorID)
		}
		return nil, fmt.Errorf("failed to get keypair: %v", err)
	}

	return &models.SSHKeypairResponse{
		ID:            keypair.ID,
		AggregatorID:  keypair.AggregatorID,
		KeyName:       keypair.KeyName,
		CloudProvider: keypair.CloudProvider,
		Region:        keypair.Region,
		PublicKey:     keypair.PublicKey,
		CreatedAt:     keypair.CreatedAt,
		UpdatedAt:     keypair.UpdatedAt,
	}, nil
}

// GetKeypairWithPrivateKey Private Key를 포함한 키페어 조회 (다운로드용)
func (s *SSHKeypairService) GetKeypairWithPrivateKey(aggregatorID string) (*models.SSHKeypairWithPrivateKey, error) {
	keypair, err := s.repo.GetKeypairByAggregatorID(aggregatorID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("keypair not found for aggregator: %s", aggregatorID)
		}
		return nil, fmt.Errorf("failed to get keypair: %v", err)
	}

	// Private Key 복호화
	privateKey, err := utils.DecryptPrivateKey(keypair.EncryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	return &models.SSHKeypairWithPrivateKey{
		SSHKeypairResponse: models.SSHKeypairResponse{
			ID:            keypair.ID,
			AggregatorID:  keypair.AggregatorID,
			KeyName:       keypair.KeyName,
			CloudProvider: keypair.CloudProvider,
			Region:        keypair.Region,
			PublicKey:     keypair.PublicKey,
			CreatedAt:     keypair.CreatedAt,
			UpdatedAt:     keypair.UpdatedAt,
		},
		PrivateKey: privateKey,
	}, nil
}

// DeleteKeypairByAggregatorID 집계자 ID로 키페어 삭제
func (s *SSHKeypairService) DeleteKeypairByAggregatorID(aggregatorID string) error {
	// 키페어 존재 확인
	_, err := s.repo.GetKeypairByAggregatorID(aggregatorID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("keypair not found for aggregator: %s", aggregatorID)
		}
		return fmt.Errorf("failed to check keypair: %v", err)
	}

	return s.repo.DeleteKeypairByAggregatorID(aggregatorID)
}

// ListKeypairsByUserID 사용자의 모든 키페어 목록 조회
func (s *SSHKeypairService) ListKeypairsByUserID(userID int64) ([]*models.SSHKeypairResponse, error) {
	keypairs, err := s.repo.ListKeypairsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list keypairs: %v", err)
	}

	responses := make([]*models.SSHKeypairResponse, len(keypairs))
	for i, kp := range keypairs {
		responses[i] = &models.SSHKeypairResponse{
			ID:            kp.ID,
			AggregatorID:  kp.AggregatorID,
			KeyName:       kp.KeyName,
			CloudProvider: kp.CloudProvider,
			Region:        kp.Region,
			PublicKey:     kp.PublicKey,
			CreatedAt:     kp.CreatedAt,
			UpdatedAt:     kp.UpdatedAt,
		}
	}

	return responses, nil
}
