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
	repo *repository.AggregatorRepository
}

// NewAggregatorHandler는 새 AggregatorHandler 인스턴스를 생성합니다
func NewAggregatorHandler(repo *repository.AggregatorRepository) *AggregatorHandler {
	return &AggregatorHandler{repo: repo}
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

// CreateAggregator godoc
// @Summary 새 Aggregator 생성
// @Description 새로운 Aggregator 인스턴스를 생성합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param aggregator body CreateAggregatorRequest true "Aggregator 생성 정보"
// @Success 201 {object} models.Aggregator
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// Aggregator 객체 생성
	aggregator := &models.Aggregator{
		ID:                    uuid.New().String(),
		UserID:               userID,
		Name:                 request.Name,
		Status:               "pending",
		Algorithm:            request.Algorithm,
		CloudProvider:        request.CloudProvider,
		Region:               request.Region,
		InstanceType:         request.InstanceType,
		ParticipantCount:     request.Participants,
		Rounds:               request.Rounds,
		CurrentRound:         0,
		EstimatedCost:        request.EstimatedCost,
		CPUSpecs:             request.CPUSpecs,
		MemorySpecs:          request.MemorySpecs,
		StorageSpecs:         request.StorageSpecs,
		Configuration:        request.Configuration,
	}

	if err := h.repo.CreateAggregator(aggregator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 생성 실패"})
		return
	}

	c.JSON(http.StatusCreated, aggregator)
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

// Request/Response structures
type CreateAggregatorRequest struct {
	Name                  string                 `json:"name" binding:"required"`
	Algorithm             string                 `json:"algorithm" binding:"required"`
	FederatedLearningID   string                 `json:"federated_learning_id" binding:"required"`
	FederatedLearningName string                 `json:"federated_learning_name"`
	CloudProvider         string                 `json:"cloud_provider" binding:"required"`
	Region                string                 `json:"region" binding:"required"`
	InstanceType          string                 `json:"instance_type" binding:"required"`
	Participants          int                    `json:"participants"`
	Rounds                int                    `json:"rounds"`
	EstimatedCost         float64                `json:"estimated_cost"`
	CPUSpecs              string                 `json:"cpu_specs"`
	MemorySpecs           string                 `json:"memory_specs"`
	StorageSpecs          string                 `json:"storage_specs"`
	Configuration         map[string]interface{} `json:"configuration"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateMetricsRequest struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	NetworkUsage float64 `json:"network_usage"`
}
