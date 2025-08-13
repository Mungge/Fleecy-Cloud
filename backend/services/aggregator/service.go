package aggregator

import (
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// AggregatorService는 Aggregator 관련 비즈니스 로직을 처리합니다
type AggregatorService struct {
	repo   *repository.AggregatorRepository
	flRepo *repository.FederatedLearningRepository
}

// NewAggregatorService는 새 AggregatorService 인스턴스를 생성합니다
func NewAggregatorService(repo *repository.AggregatorRepository, flRepo *repository.FederatedLearningRepository) *AggregatorService {
	return &AggregatorService{
		repo:   repo,
		flRepo: flRepo,
	}
}

// CreateAggregatorInput Aggregator 생성 입력
type CreateAggregatorInput struct {
	Name         string
	Algorithm    string
	Region       string
	Storage      string
	InstanceType string
	UserID       int64
}

// CreateAggregatorResult Aggregator 생성 결과
type CreateAggregatorResult struct {
	AggregatorID    string
	Status          string
	TerraformStatus string
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
	// Aggregator 생성
	aggregator := &models.Aggregator{
		ID:            uuid.New().String(),
		UserID:        input.UserID,
		Name:          input.Name,
		Status:        "creating",
		Algorithm:     input.Algorithm,
		CloudProvider: "openstack",
		Region:        input.Region,
		InstanceType:  input.InstanceType,
		StorageSpecs:  input.Storage + "GB",
	}

	// DB에 저장
	if err := s.repo.CreateAggregator(aggregator); err != nil {
		return nil, err
	}

	// Terraform 배포 시작 (비동기)
	go func() {
		if err := s.deployWithTerraform(aggregator); err != nil {
			aggregator.Status = "failed"
			s.repo.UpdateAggregator(aggregator)
			return
		}
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

// deployWithTerraform은 Terraform을 사용하여 인프라를 배포합니다
func (s *AggregatorService) deployWithTerraform(aggregator *models.Aggregator) error {
	// Terraform 설정 생성
	config := utils.TerraformConfig{
		Region:       aggregator.Region,
		InstanceType: aggregator.InstanceType,
		ProjectName:  "fleecy-aggregator",
		Environment:  "production",
	}
	
	// Terraform 작업공간 생성
	workspaceDir, err := utils.CreateTerraformWorkspace(aggregator.ID, config)
	if err != nil {
		return err
	}
	
	// Terraform 배포 실행
	result, err := utils.DeployWithTerraform(workspaceDir)
	if err != nil {
		return err
	}
	
	// 배포 결과를 aggregator에 저장
	aggregator.InstanceID = result.InstanceID
	aggregator.Status = "running"
	if aggregator.Configuration == nil {
		aggregator.Configuration = make(map[string]interface{})
	}
	aggregator.Configuration["public_ip"] = result.PublicIP
	aggregator.Configuration["private_ip"] = result.PrivateIP
	aggregator.Configuration["workspace_dir"] = result.WorkspaceDir
	
	return nil
}

//GetAggregatorsByUser는 사용자의 모든 Aggregator를 조회합니다
func (s *AggregatorService) GetAggregatorsByUser(userID int64) ([]*models.Aggregator, error) {
	return s.repo.GetAggregatorsByUserID(userID)
}

// GetAggregator