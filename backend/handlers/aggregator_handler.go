package handlers

import (
	"fmt"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// AggregatorHandler는 Aggregator 관련 API 핸들러입니다
type AggregatorHandler struct {
	repo   *repository.AggregatorRepository
	flRepo *repository.FederatedLearningRepository
	optimizationService *services.OptimizationService
}

// NewAggregatorHandler는 새 AggregatorHandler 인스턴스를 생성합니다
func NewAggregatorHandler(repo *repository.AggregatorRepository, flRepo *repository.FederatedLearningRepository) *AggregatorHandler {
	return &AggregatorHandler{
		repo:   repo,
		flRepo: flRepo,
		optimizationService: services.NewOptimizationService(),
	}
}

// OptimizeAggregatorPlacement godoc
// @Summary 집계자 배치 최적화
// @Description NSGA-II 알고리즘을 사용하여 사용자 제약사항에 맞는 최적의 집계자 배치 위치를 찾습니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param request body services.OptimizationRequest true "최적화 요청 정보"
// @Success 200 {object} map[string]interface{} "data: services.OptimizationResponse"
// @Failure 400 {object} map[string]string "error: 잘못된 요청"
// @Failure 401 {object} map[string]string "error: 인증 필요"
// @Failure 500 {object} map[string]string "error: 서버 오류"
// @Router /api/aggregator-placement-optimization [post]
func (h *AggregatorHandler) OptimizeAggregatorPlacement(c *gin.Context) {
	// 1. 사용자 인증 확인 (선택적 - 필요에 따라)
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	// 2. 요청 데이터 파싱
	var request services.OptimizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "잘못된 요청 형식입니다: " + err.Error(),
		})
		return
	}

	// 3. 요청 데이터 유효성 검증
	if err := h.validateOptimizationRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "요청 데이터 유효성 검증 실패: " + err.Error(),
		})
		return
	}

	// 4. Python 환경 및 스크립트 존재 여부 확인
	if err := h.optimizationService.ValidatePythonEnvironment(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Python 환경 오류: " + err.Error(),
		})
		return
	}

	if err := h.optimizationService.ValidatePythonScript(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Python 스크립트 오류: " + err.Error(),
		})
		return
	}

	// 5. 최적화 실행
	result, err := h.optimizationService.RunOptimization(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "집계자 배치 최적화 실행 실패: " + err.Error(),
		})
		return
	}

	// 6. 성공 응답 반환
	c.JSON(http.StatusOK, gin.H{
		"message": "집계자 배치 최적화가 성공적으로 완료되었습니다",
		"data":    result,
		"userID":  userID, // 디버깅용 (필요에 따라 제거)
	})
}

// validateOptimizationRequest 요청 데이터 유효성 검증
func (h *AggregatorHandler) validateOptimizationRequest(request *services.OptimizationRequest) error {
	// 1. 연합학습 정보 검증
	if request.FederatedLearning.Name == "" {
		return fmt.Errorf("연합학습 이름이 필요합니다")
	}

	if request.FederatedLearning.Algorithm == "" {
		return fmt.Errorf("집계 알고리즘이 필요합니다")
	}

	if request.FederatedLearning.Rounds <= 0 {
		return fmt.Errorf("라운드 수는 1 이상이어야 합니다")
	}

	// 2. 참여자 정보 검증
	if len(request.FederatedLearning.Participants) == 0 {
		return fmt.Errorf("최소 1명 이상의 참여자가 필요합니다")
	}

	for i, participant := range request.FederatedLearning.Participants {
		if participant.ID == "" {
			return fmt.Errorf("참여자 %d의 ID가 필요합니다", i+1)
		}
		if participant.Name == "" {
			return fmt.Errorf("참여자 %d의 이름이 필요합니다", i+1)
		}
		if participant.OpenstackEndpoint == "" {
			return fmt.Errorf("참여자 %d의 OpenStack 엔드포인트가 필요합니다", i+1)
		}
		// 간단한 URL 형식 검증
		if !strings.HasPrefix(participant.OpenstackEndpoint, "http://") && 
		   !strings.HasPrefix(participant.OpenstackEndpoint, "https://") {
			return fmt.Errorf("참여자 %d의 OpenStack 엔드포인트가 올바른 URL 형식이 아닙니다", i+1)
		}
	}

	// 3. 제약사항 검증
	if request.Constraints.MaxBudget <= 0 {
		return fmt.Errorf("최대 예산은 0보다 커야 합니다")
	}

	if request.Constraints.MaxLatency <= 0 {
		return fmt.Errorf("최대 지연시간은 0보다 커야 합니다")
	}

	// 합리적인 범위 체크
	if request.Constraints.MaxBudget > 10000000 { // 1천만원
		return fmt.Errorf("최대 예산이 너무 큽니다 (최대: 10,000,000원)")
	}

	if request.Constraints.MaxLatency > 10000 { // 10초
		return fmt.Errorf("최대 지연시간이 너무 큽니다 (최대: 10,000ms)")
	}

	return nil
}


// GetAggregators godoc
// @Summary Aggregator 목록 조회
// @Description 사용자의 모든 Aggregator 인스턴스를 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Success 200 {array} models.Aggregator
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators [get]
func (h *AggregatorHandler) GetAggregators(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregators, err := h.repo.GetAggregatorsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 목록 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, aggregators)
}

// GetAggregator godoc
// @Summary 특정 Aggregator 조회
// @Description ID로 특정 Aggregator를 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {object} models.Aggregator
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id} [get]
func (h *AggregatorHandler) GetAggregator(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	id := c.Param("id")
	aggregator, err := h.repo.GetAggregatorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	c.JSON(http.StatusOK, aggregator)
}

// UpdateAggregatorStatus godoc
// @Summary Aggregator 상태 업데이트
// @Description Aggregator의 상태를 업데이트합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Param status body UpdateStatusRequest true "상태 정보"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/status [put]
func (h *AggregatorHandler) UpdateAggregatorStatus(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	id := c.Param("id")
	
	// 권한 확인
	aggregator, err := h.repo.GetAggregatorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}
	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	var request UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	if err := h.repo.UpdateAggregatorStatus(id, request.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "상태 업데이트 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "상태가 업데이트되었습니다"})
}

// UpdateAggregatorMetrics godoc
// @Summary Aggregator 메트릭 업데이트
// @Description Aggregator의 실시간 메트릭을 업데이트합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Param metrics body UpdateMetricsRequest true "메트릭 정보"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/metrics [put]
func (h *AggregatorHandler) UpdateAggregatorMetrics(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	id := c.Param("id")
	
	// 권한 확인
	aggregator, err := h.repo.GetAggregatorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}
	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	var request UpdateMetricsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	err = h.repo.UpdateAggregatorMetrics(id, request.CPUUsage, request.MemoryUsage, request.NetworkUsage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "메트릭 업데이트 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "메트릭이 업데이트되었습니다"})
}

// GetTrainingHistory godoc
// @Summary 학습 히스토리 조회
// @Description Aggregator의 학습 히스토리를 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {array} models.TrainingRound
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/training-history [get]
func (h *AggregatorHandler) GetTrainingHistory(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	id := c.Param("id")
	
	// 권한 확인
	aggregator, err := h.repo.GetAggregatorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}
	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	rounds, err := h.repo.GetTrainingRoundsByAggregatorID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "학습 히스토리 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, rounds)
}

// GetAggregatorStats godoc
// @Summary Aggregator 통계 조회
// @Description 사용자의 Aggregator 통계를 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/stats [get]
func (h *AggregatorHandler) GetAggregatorStats(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	stats, err := h.repo.GetAggregatorStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DeleteAggregator godoc
// @Summary Aggregator 삭제
// @Description Aggregator를 삭제합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id} [delete]
func (h *AggregatorHandler) DeleteAggregator(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	id := c.Param("id")
	
	// 권한 확인
	aggregator, err := h.repo.GetAggregatorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}
	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	if err := h.repo.DeleteAggregator(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 삭제 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aggregator가 삭제되었습니다"})
}


// CreateAggregator godoc
// @Summary 새 Aggregator 생성
// @Description 새로운 Aggregator 인스턴스를 생성하고 Terraform으로 인프라를 배포합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param aggregator body CreateAggregatorRequest true "Aggregator 생성 정보"
// @Success 201 {object} CreateAggregatorResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators [post]
func (h *AggregatorHandler) CreateAggregator(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	var request CreateAggregatorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다: " + err.Error()})
		return
	}

	// Aggregator 생성
	aggregator := &models.Aggregator{
		ID:            uuid.New().String(),
		UserID:       userID,
		Name:         request.Name,
		Status:       "creating",
		Algorithm:    request.Algorithm,
		CloudProvider: "openstack",
		Region:       request.Region,
		InstanceType: request.InstanceType,
		StorageSpecs: request.Storage + "GB",
	}

	// DB에 저장
	if err := h.repo.CreateAggregator(aggregator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 생성 실패: " + err.Error()})
		return
	}

	// Terraform 배포 시작 (비동기)
	go func() {
		if err := h.deployWithTerraform(aggregator); err != nil {
			aggregator.Status = "failed"
			h.repo.UpdateAggregator(aggregator)
			return
		}
		aggregator.Status = "running"
		h.repo.UpdateAggregator(aggregator)
	}()

	// 응답 반환
	response := CreateAggregatorResponse{
		AggregatorID:    aggregator.ID,
		Status:          "creating",
		TerraformStatus: "deploying",
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// deployWithTerraform은 Terraform을 사용하여 인프라를 배포합니다
func (h *AggregatorHandler) deployWithTerraform(aggregator *models.Aggregator) error {
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


// Request/Response structures
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateMetricsRequest struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	NetworkUsage float64 `json:"network_usage"`
}

// Aggregator 생성 요청 구조
type CreateAggregatorRequest struct {
	Name         string `json:"name" binding:"required"`
	Algorithm    string `json:"algorithm" binding:"required"`
	Region       string `json:"region" binding:"required"`
	Storage      string `json:"storage" binding:"required"`
	InstanceType string `json:"instanceType" binding:"required"`
}

// Aggregator 생성 응답 구조
type CreateAggregatorResponse struct {
	AggregatorID    string `json:"aggregatorId"`
	Status          string `json:"status"`
	TerraformStatus string `json:"terraformStatus,omitempty"`
}

// 프론트엔드에서 오는 연합학습 데이터 구조 (통합 생성용)
type FederatedLearningData struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ModelType    string `json:"modelType"`
	Algorithm    string `json:"algorithm"`
	Rounds       int    `json:"rounds"`
	Participants []struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		Status            string `json:"status"`
		OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
	} `json:"participants"`
	ModelFileName string `json:"modelFileName,omitempty"`
}

// 프론트엔드에서 오는 집계자 설정 구조
type AggregatorConfigData struct {
	Region       string `json:"region"`
	Storage      string `json:"storage"`
	InstanceType string `json:"instanceType"`
}

type OptimizationRequest struct {
    FederatedLearning struct {
        Name         string        `json:"name"`
        Description  string        `json:"description"`
        Algorithm    string        `json:"algorithm"`
        Rounds       int           `json:"rounds"`
        Participants []Participant `json:"participants"`
    } `json:"federatedLearning"`
    Constraints struct {
        MaxBudget  int `json:"maxBudget"`
        MaxLatency int `json:"maxLatency"`
    } `json:"constraints"`
}

type Participant struct {
    ID               string `json:"id"`
    Name             string `json:"name"`
    OpenstackEndpoint string `json:"openstack_endpoint"`
}