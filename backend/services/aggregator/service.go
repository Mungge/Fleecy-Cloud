package aggregator

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	repo            *repository.AggregatorRepository
	flRepo          *repository.FederatedLearningRepository
	sshKeypairRepo  *repository.SSHKeypairRepository
	cloudRepo       *repository.CloudRepository
	progressTracker *SSEProgressTracker
}

// NewAggregatorService는 새 AggregatorService 인스턴스를 생성합니다
func NewAggregatorService(repo *repository.AggregatorRepository, flRepo *repository.FederatedLearningRepository, sshKeypairRepo *repository.SSHKeypairRepository, cloudRepo *repository.CloudRepository) *AggregatorService {
	return &AggregatorService{
		repo:            repo,
		flRepo:          flRepo,
		sshKeypairRepo:  sshKeypairRepo,
		cloudRepo:       cloudRepo,
		progressTracker: NewWebSocketProgressTracker(),
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

// CreateAggregator는 새로운 Aggregator를 생성합니다 (기본 컨텍스트 사용)
func (s *AggregatorService) CreateAggregator(input CreateAggregatorInput) (*CreateAggregatorResult, error) {
	return s.CreateAggregatorWithContext(context.Background(), input)
}

// CreateAggregatorWithContext는 컨텍스트를 지원하는 새로운 Aggregator 생성 메서드입니다
func (s *AggregatorService) CreateAggregatorWithContext(ctx context.Context, input CreateAggregatorInput) (*CreateAggregatorResult, error) {

	// 동일한 사용자의 동일한 이름 집계자가 이미 존재하는지 확인
	existingAggregators, err := s.repo.GetAggregatorsByUserID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("기존 집계자 조회 실패: %w", err)
	}

	for _, existing := range existingAggregators {
		if existing.Name == input.Name && (existing.Status == "creating" || existing.Status == "running") {
			return nil, fmt.Errorf("동일한 이름의 집계자가 이미 존재합니다: %s", input.Name)
		}
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

	// DB에 저장 (creating 상태로)
	if err := s.repo.CreateAggregator(aggregator); err != nil {
		return nil, err
	}

	// Terraform 배포 시작 (동기) - 사용자가 결과를 즉시 확인 가능
	log.Printf("Starting synchronous Terraform deployment for aggregator %s", aggregator.ID)

	if err := s.deployWithTerraformContext(ctx, aggregator); err != nil {
		log.Printf("Terraform deployment failed for aggregator %s: %v", aggregator.ID, err)

		// 배포 실패 시 상태 업데이트
		aggregator.Status = "failed"
		if updateErr := s.repo.UpdateAggregatorStatus(aggregator.ID, "failed"); updateErr != nil {
			log.Printf("Failed to update aggregator status to failed: %v", updateErr)
		}

		// 사용자 친화적인 에러 메시지 생성
		var userMessage string
		if strings.Contains(err.Error(), "active AWS connection not found") {
			userMessage = "AWS 클라우드 연결이 설정되지 않았습니다. 먼저 클라우드 인증 정보를 등록해주세요."
		} else if strings.Contains(err.Error(), "active GCP connection not found") {
			userMessage = "GCP 클라우드 연결이 설정되지 않았습니다. 먼저 클라우드 인증 정보를 등록해주세요."
		} else if strings.Contains(err.Error(), "AWS credentials") {
			userMessage = "AWS 자격증명이 올바르지 않습니다. 클라우드 인증 정보를 다시 확인해주세요."
		} else {
			userMessage = fmt.Sprintf("집계자 배포 실패: %v", err)
		}

		return nil, fmt.Errorf("%s", userMessage)
	}

	// 배포 성공 시 상태 업데이트
	log.Printf("Terraform deployment successful for aggregator %s", aggregator.ID)
	aggregator.Status = "running"
	if updateErr := s.repo.UpdateAggregatorStatus(aggregator.ID, "running"); updateErr != nil {
		log.Printf("Failed to update aggregator status to running: %v", updateErr)
		// 상태 업데이트 실패해도 배포는 성공했으므로 계속 진행
	}

	// 결과 반환
	result := &CreateAggregatorResult{
		AggregatorID:    aggregator.ID,
		Status:          aggregator.Status, // 실제 상태 반환
		TerraformStatus: "completed",       // 배포 완료
	}

	return result, nil
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

// deployWithTerraformContext는 컨텍스트를 지원하는 Terraform 배포 메서드입니다
func (s *AggregatorService) deployWithTerraformContext(ctx context.Context, aggregator *models.Aggregator) error {
	log.Printf("[%s] 1/5 클라우드 연결 정보 조회 중...", aggregator.ID)
	s.progressTracker.SendProgress(aggregator.ID, 1, "클라우드 연결 정보 조회 중...")

	// 컨텍스트 취소 확인
	if ctx.Err() != nil {
		s.progressTracker.SendError(aggregator.ID, 1, "배포 취소됨", ctx.Err())
		return fmt.Errorf("deployment cancelled: %v", ctx.Err())
	}

	// 사용자의 클라우드 연결 정보 가져오기
	cloudConnections, err := s.cloudRepo.GetCloudConnectionsByUserID(aggregator.UserID)
	if err != nil {
		return fmt.Errorf("failed to get cloud connections: %v", err)
	}

	// 해당 클라우드 제공업체의 연결 찾기 (대소문자 구분 없이)
	var cloudConn *models.CloudConnection
	for _, conn := range cloudConnections {
		if strings.EqualFold(conn.Provider, aggregator.CloudProvider) && conn.Status == "active" {
			cloudConn = conn
			break
		}
	}

	if cloudConn == nil {
		return fmt.Errorf("no %s cloud connection found for user %d", aggregator.CloudProvider, aggregator.UserID)
	}

	log.Printf("[%s] 2/5 클라우드 자격증명 파싱 중...", aggregator.ID)
	s.progressTracker.SendProgress(aggregator.ID, 2, "클라우드 자격증명 파싱 중...")

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

	log.Printf("[%s] 3/5 SSH 키페어 생성/조회 중...", aggregator.ID)
	s.progressTracker.SendProgress(aggregator.ID, 3, "SSH 키페어 생성/조회 중...")

	// 컨텍스트 취소 확인
	if ctx.Err() != nil {
		s.progressTracker.SendError(aggregator.ID, 3, "배포 취소됨", ctx.Err())
		return fmt.Errorf("deployment cancelled: %v", ctx.Err())
	}

	// 클라우드 키페어 서비스 초기화
	keypairService := services.NewCloudKeypairService(s.cloudRepo)

	// 클라우드별 키페어 생성/조회
	keyName := fmt.Sprintf("%s-%s-keypair", aggregator.ProjectName, aggregator.ID)

	var privateKey string

	switch strings.ToLower(aggregator.CloudProvider) {
	case "aws":
		log.Printf("Creating AWS keypair for aggregator %s in region %s", aggregator.ID, aggregator.Region)
		keypairInfo, err := keypairService.GetOrCreateAWSKeypair(aggregator.UserID, aggregator.Region, keyName)
		if err != nil {
			return fmt.Errorf("failed to get or create AWS keypair: %v", err)
		}
		privateKey = keypairInfo.PrivateKey
		log.Printf("AWS keypair created successfully: %s", keypairInfo.KeyName)
	case "gcp":
		log.Printf("Creating GCP keypair for aggregator %s", aggregator.ID)
		// GCP의 경우 SSH 키를 직접 생성해서 전달
		keyPair, err := utils.GenerateSSHKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate SSH keypair for GCP: %v", err)
		}

		_, err = keypairService.GetOrCreateGCPKeypair(aggregator.UserID, keyName, keyPair.PublicKey, keyPair.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to get or create GCP keypair: %v", err)
		}
		privateKey = keyPair.PrivateKey
		log.Printf("GCP keypair created successfully: %s", keyName)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", aggregator.CloudProvider)
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

	var projectID string
    if aggregator.CloudProvider == "gcp" {
        // GCP 자격증명에서 project_id 추출
        var credentials map[string]interface{}
        if err := json.Unmarshal(cloudConn.CredentialFile, &credentials); err != nil {
            return fmt.Errorf("failed to parse GCP credentials: %v", err)
        }
        if id, ok := credentials["project_id"].(string); ok {
            projectID = id
        } else {
            return fmt.Errorf("project_id not found in GCP credentials")
        }
    }

	log.Printf("[%s] 4/5 Terraform 워크스페이스 생성 중...", aggregator.ID)
	s.progressTracker.SendProgress(aggregator.ID, 4, "Terraform 워크스페이스 생성 중...")

	// 컨텍스트 취소 확인
	if ctx.Err() != nil {
		s.progressTracker.SendError(aggregator.ID, 4, "배포 취소됨", ctx.Err())
		return fmt.Errorf("deployment cancelled: %v", ctx.Err())
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
		AWSAccessKey:  awsAccessKey,
		AWSSecretKey:  awsSecretKey,
		ProjectID:     projectID,
		GCPServiceAccountKey: string(cloudConn.CredentialFile), // JSON을 문자열로 변환
	}

	// Terraform 작업공간 생성
	workspaceDir, err := utils.CreateTerraformWorkspace(aggregator.ID, config)
	if err != nil {
		return err
	}

	s.progressTracker.SendProgress(aggregator.ID, 5, "Terraform 배포 실행 중... (시간이 소요될 수 있습니다)")

	// 컨텍스트 취소 확인
	if ctx.Err() != nil {
		s.progressTracker.SendError(aggregator.ID, 5, "배포 취소됨", ctx.Err())
		return fmt.Errorf("deployment cancelled: %v", ctx.Err())
	}

	// Terraform 배포 실행 (terraform-exec 사용, 컨텍스트 지원)
	result, err := utils.DeployWithTerraformContext(ctx, workspaceDir)

	// 배포 완료 후 workspace 디렉토리 정리 (성공/실패 상관없이)
	defer func() {
		if cleanupErr := os.RemoveAll(workspaceDir); cleanupErr != nil {
			log.Printf("Failed to cleanup workspace directory %s: %v", workspaceDir, cleanupErr)
		} else {
			log.Printf("Cleaned up workspace directory: %s", workspaceDir)
		}
	}()

	if err != nil {
		s.progressTracker.SendError(aggregator.ID, 5, "Terraform 배포 실패", err)
		return err
	}

	// 배포 성공
	s.progressTracker.SendSuccess(aggregator.ID, "집계자 배포가 성공적으로 완료되었습니다")

	// Terraform 결과 로깅
	log.Printf("[%s] Terraform deployment result: InstanceID=%s, PublicIP=%s, PrivateIP=%s", 
		aggregator.ID, result.InstanceID, result.PublicIP, result.PrivateIP)

	// 배포 결과를 aggregator에 저장
	aggregator.InstanceID = result.InstanceID
	aggregator.PublicIP = result.PublicIP
	aggregator.PrivateIP = result.PrivateIP
	aggregator.Status = "running"

	// IP 정보를 데이터베이스에 업데이트
	if err := s.repo.UpdateAggregatorIPInfo(aggregator.ID, result.InstanceID, result.PublicIP, result.PrivateIP); err != nil {
		log.Printf("Failed to update aggregator IP info: %v", err)
		// IP 정보 업데이트 실패해도 배포는 성공했으므로 계속 진행
	} else {
		log.Printf("Updated aggregator %s with IP info: Public=%s, Private=%s", aggregator.ID, result.PublicIP, result.PrivateIP)
	}

	return nil
}

// HandleWebSocketProgress WebSocket 진행 상황 연결 처리
func (s *AggregatorService) HandleWebSocketProgress(w http.ResponseWriter, r *http.Request, aggregatorID string) {
	s.progressTracker.HandleWebSocket(w, r, aggregatorID)
}
