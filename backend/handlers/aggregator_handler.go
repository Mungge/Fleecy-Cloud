package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// AggregatorHandler는 Aggregator 관련 API 핸들러입니다
type AggregatorHandler struct {
	repo   *repository.AggregatorRepository
	flRepo *repository.FederatedLearningRepository
}

// NewAggregatorHandler는 새 AggregatorHandler 인스턴스를 생성합니다
func NewAggregatorHandler(repo *repository.AggregatorRepository, flRepo *repository.FederatedLearningRepository) *AggregatorHandler {
	return &AggregatorHandler{
		repo:   repo,
		flRepo: flRepo,
	}
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
