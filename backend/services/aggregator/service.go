package aggregator

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// AggregatorService는 Aggregator 관련 비즈니스 로직을 처리합니다
type AggregatorService struct {
	repo           *repository.AggregatorRepository
	flRepo         *repository.FederatedLearningRepository
	sshKeypairRepo *repository.SSHKeypairRepository
	cloudRepo      *repository.CloudRepository
}

// NewAggregatorService는 새 AggregatorService 인스턴스를 생성합니다
func NewAggregatorService(repo *repository.AggregatorRepository, flRepo *repository.FederatedLearningRepository, sshKeypairRepo *repository.SSHKeypairRepository, cloudRepo *repository.CloudRepository) *AggregatorService {
	return &AggregatorService{
		repo:           repo,
		flRepo:         flRepo,
		sshKeypairRepo: sshKeypairRepo,
		cloudRepo:      cloudRepo,
	}
}

// CreateAggregatorInput Aggregator 생성 입력
type CreateAggregatorInput struct {
	Name          string `json:"name" validate:"required"`
	Algorithm     string `json:"algorithm" validate:"required"`
	Storage       string `json:"storage" validate:"required"`
	UserID        int64  `json:"user_id" validate:"required"`
	CloudProvider string `json:"cloud_provider" validate:"required,oneof=aws gcp"`

	// 공통 필드
	ProjectName  string `json:"project_name" validate:"required"`
	Region       string `json:"region" validate:"required"`
	Zone         string `json:"zone" validate:"required"`
	InstanceType string `json:"instance_type" validate:"required"`

	// GCP 전용 (선택적)
	ProjectID string `json:"project_id,omitempty"`
}

// CreateAggregatorResult Aggregator 생성 결과
type CreateAggregatorResult struct {
	AggregatorID    string `json:"aggregator_id"`
	Status          string `json:"status"`
	TerraformStatus string `json:"terraform_status"`
}

// OptimizationRequest 최적화 요청 (기존 services.OptimizationRequest 대체)
type OptimizationRequest struct {
	FederatedLearning struct {
		Name         string        `json:"name"`
		Description  string        `json:"description"`
		Algorithm    string        `json:"algorithm"`
		Rounds       int           `json:"rounds"`
		Participants []Participant `json:"participants"`
	} `json:"federatedLearning"`
	AggregatorConfig struct {
		MaxBudget  int `json:"maxBudget"`
		MaxLatency int `json:"maxLatency"`
	} `json:"aggregatorConfig"`
}

// Participant 참여자 정보
type Participant struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Region            string `json:"region,omitempty"`
	OpenstackEndpoint string `json:"openstack_endpoint"`
}

// OptimizationService 인터페이스 (기존 services.OptimizationService 대체)
type OptimizationService interface {
	ValidatePythonEnvironment() error
	ValidatePythonScript() error
	RunOptimization(request OptimizationRequest) (interface{}, error)
}

// CreateAggregator는 새로운 Aggregator를 생성합니다
func (s *AggregatorService) CreateAggregator(input CreateAggregatorInput) (*CreateAggregatorResult, error) {
	// 입력 검증
	if err := s.validateInput(input); err != nil {
		return nil, err
	}

	// Aggregator 생성
	aggregator := &models.Aggregator{
		ID:            uuid.New().String(),
		UserID:        input.UserID,
		Name:          input.Name,
		Status:        "creating",
		Algorithm:     input.Algorithm,
		CloudProvider: input.CloudProvider,
		ProjectName:   input.ProjectName,
		Region:        input.Region,
		Zone:          input.Zone,
		InstanceType:  input.InstanceType,
		StorageSpecs:  input.Storage + "GB",
	}

	// GCP인 경우 ProjectID 설정
	if input.CloudProvider == "gcp" && input.ProjectID != "" {
		aggregator.ProjectID = &input.ProjectID
	}

	// DB에 저장
	if err := s.repo.CreateAggregator(aggregator); err != nil {
		return nil, err
	}

	// Terraform 배포 시작 (비동기) - terraform-exec 기반
	go func() {
		if err := s.deployWithTerraform(aggregator); err != nil {
			log.Printf("Terraform deployment failed for aggregator %s: %v", aggregator.ID, err)
			aggregator.Status = "failed"
			s.repo.UpdateAggregator(aggregator)
			return
		}
		log.Printf("Terraform deployment successful for aggregator %s", aggregator.ID)
		aggregator.Status = "running"
		s.repo.UpdateAggregator(aggregator)
	}()

	// 결과 반환
	result := &CreateAggregatorResult{
		AggregatorID:    aggregator.ID,
		Status:          "creating",
		TerraformStatus: "deploying",
	}

	return result, nil
}

// validateInput 입력값 검증
func (s *AggregatorService) validateInput(input CreateAggregatorInput) error {
	if input.CloudProvider == "gcp" && input.ProjectID == "" {
		return ErrGCPNeedsProjectID
	}
	return nil
}

// GetAggregatorByID는 ID로 Aggregator를 조회하고 권한을 확인합니다
func (s *AggregatorService) GetAggregatorByID(id string, userID int64) (*models.Aggregator, error) {
	aggregator, err := s.repo.GetAggregatorByID(id)
	if err != nil {
		return nil, err
	}

	if aggregator == nil || aggregator.UserID != userID {
		return nil, nil // 권한 없음 또는 존재하지 않음
	}

	return aggregator, nil
}

// DeleteAggregator는 Aggregator를 삭제합니다
func (s *AggregatorService) DeleteAggregator(id string, userID int64) error {
	// 권한 확인
	aggregator, err := s.GetAggregatorByID(id, userID)
	if err != nil {
		return err
	}
	if aggregator == nil {
		return ErrAggregatorNotFound
	}

	return s.repo.DeleteAggregator(id)
}

// GetAggregatorsByUser는 사용자의 모든 Aggregator를 조회합니다
func (s *AggregatorService) GetAggregatorsByUser(userID int64) ([]*models.Aggregator, error) {
	return s.repo.GetAggregatorsByUserID(userID)
}

// deployWithTerraform은 Terraform을 사용하여 Aggregator를 배포합니다
func (s *AggregatorService) deployWithTerraform(aggregator *models.Aggregator) error {
	// 사용자의 클라우드 연결 정보 가져오기
	cloudConnections, err := s.cloudRepo.GetCloudConnectionsByUserID(aggregator.UserID)
	if err != nil {
		return fmt.Errorf("failed to get cloud connections: %v", err)
	}

	// 해당 클라우드 제공업체의 연결 찾기 (대소문자 구분 없이)
	var cloudConn *models.CloudConnection
	for _, conn := range cloudConnections {
		if strings.EqualFold(conn.Provider, aggregator.CloudProvider) {
			cloudConn = conn
			break
		}
	}

	if cloudConn == nil {
		return fmt.Errorf("no %s cloud connection found for user %d", aggregator.CloudProvider, aggregator.UserID)
	}

	// 자격증명 파싱
	var awsAccessKey, awsSecretKey string
	if aggregator.CloudProvider == "aws" {
		// BOM 및 기타 불필요한 문자 제거
		credentialData := bytes.TrimPrefix(cloudConn.CredentialFile, []byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM 제거
		credentialData = bytes.TrimSpace(credentialData)                                       // 앞뒤 공백 제거

		// JSON 형식으로 먼저 시도
		var credentials map[string]interface{}
		if err := json.Unmarshal(credentialData, &credentials); err != nil {
			// JSON 파싱 실패 시 CSV 형식으로 시도
			log.Printf("JSON parsing failed, trying CSV format")

			reader := csv.NewReader(bytes.NewReader(credentialData))
			records, csvErr := reader.ReadAll()
			if csvErr != nil {
				log.Printf("Both JSON and CSV parsing failed. Data: %s", string(credentialData))
				return fmt.Errorf("failed to parse AWS credentials as JSON or CSV: JSON error: %v, CSV error: %v", err, csvErr)
			}

			// CSV에서 자격증명 추출 (첫 번째 데이터 행 사용)
			if len(records) >= 2 && len(records[1]) >= 2 {
				awsAccessKey = strings.TrimSpace(records[1][0])
				awsSecretKey = strings.TrimSpace(records[1][1])
			} else if len(records) >= 1 && len(records[0]) >= 2 {
				// 헤더 없이 바로 데이터인 경우
				awsAccessKey = strings.TrimSpace(records[0][0])
				awsSecretKey = strings.TrimSpace(records[0][1])
			} else {
				return fmt.Errorf("invalid CSV format: insufficient data")
			}
		} else {
			// JSON 파싱 성공
			if accessKey, ok := credentials["access_key_id"].(string); ok {
				awsAccessKey = accessKey
			}
			if secretKey, ok := credentials["secret_access_key"].(string); ok {
				awsSecretKey = secretKey
			}
		}

		if awsAccessKey == "" || awsSecretKey == "" {
			return fmt.Errorf("invalid AWS credentials: access key or secret key is empty")
		}
	}

	// 클라우드 키페어 서비스 초기화
	keypairService := services.NewCloudKeypairService(s.cloudRepo)

	// 클라우드별 키페어 생성/조회
	keyName := fmt.Sprintf("%s-%s-keypair", aggregator.ProjectName, aggregator.ID)

	var privateKey string

	switch aggregator.CloudProvider {
	case "aws":
		keypairInfo, err := keypairService.GetOrCreateAWSKeypair(aggregator.UserID, aggregator.Region, keyName)
		if err != nil {
			return fmt.Errorf("failed to get or create AWS keypair: %v", err)
		}
		privateKey = keypairInfo.PrivateKey
	case "gcp":
		// GCP의 경우 SSH 키를 직접 생성해서 전달
		keyPair, err := utils.GenerateSSHKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate SSH keypair for GCP: %v", err)
		}

		projectID := ""
		if aggregator.ProjectID != nil {
			projectID = *aggregator.ProjectID
		}

		_, err = keypairService.GetOrCreateGCPKeypair(aggregator.UserID, projectID, keyName, keyPair.PublicKey, keyPair.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to get or create GCP keypair: %v", err)
		}
		privateKey = keyPair.PrivateKey
	}

	// SSH 키페어를 DB에 암호화 저장 (Private Key가 있는 경우만)
	if privateKey != "" {
		sshKeypairService := services.NewSSHKeypairService(s.sshKeypairRepo)
		publicKey := "" // TODO: 공개키도 저장하려면 keypairInfo에서 가져오기

		_, err := sshKeypairService.SaveKeypair(
			aggregator.ID,
			keyName,
			aggregator.CloudProvider,
			aggregator.Region,
			publicKey,
			privateKey,
		)
		if err != nil {
			return fmt.Errorf("failed to save keypair to database: %v", err)
		}
	}

	// Terraform 설정 생성
	config := utils.TerraformConfig{
		CloudProvider: aggregator.CloudProvider,
		ProjectName:   aggregator.ProjectName,
		Region:        aggregator.Region,
		Zone:          aggregator.Zone,
		InstanceType:  aggregator.InstanceType,
		Environment:   "production",
		StorageSpecs:  aggregator.StorageSpecs,
		AggregatorID:  aggregator.ID,
		Algorithm:     aggregator.Algorithm,
		ProjectID:     aggregator.ProjectID, // GCP인 경우만 값이 있음
		AWSAccessKey:  awsAccessKey,
		AWSSecretKey:  awsSecretKey,
	}

	// Terraform 작업공간 생성
	workspaceDir, err := utils.CreateTerraformWorkspace(aggregator.ID, config)
	if err != nil {
		return err
	}

	// Terraform 배포 실행 (terraform-exec 사용)
	result, err := utils.DeployWithTerraform(workspaceDir)

	// 배포 완료 후 workspace 디렉토리 정리 (성공/실패 상관없이)
	defer func() {
		if cleanupErr := os.RemoveAll(workspaceDir); cleanupErr != nil {
			log.Printf("Failed to cleanup workspace directory %s: %v", workspaceDir, cleanupErr)
		} else {
			log.Printf("Cleaned up workspace directory: %s", workspaceDir)
		}
	}()

	if err != nil {
		return err
	}

	// 배포 결과를 aggregator에 저장
	aggregator.InstanceID = result.InstanceID
	aggregator.Status = "running"

	return nil
}
