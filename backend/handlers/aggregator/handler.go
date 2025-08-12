package aggregator

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/Mungge/Fleecy-Cloud/utils"
	aggregatorvalidator "github.com/Mungge/Fleecy-Cloud/validators/aggregator"
)

// AggregatorHandler는 Aggregator 관련 API 핸들러입니다
type AggregatorHandler struct {
	aggregatorService   *aggregator.AggregatorService
	metricsService      *aggregator.AggregatorMetricsService
	trainingService     *aggregator.AggregatorTrainingService
	optimizationService aggregator.OptimizationService
}

// NewAggregatorHandler는 새 AggregatorHandler 인스턴스를 생성합니다
func NewAggregatorHandler(
	aggregatorService *aggregator.AggregatorService,
	metricsService *aggregator.AggregatorMetricsService,
	trainingService *aggregator.AggregatorTrainingService,
	optimizationService aggregator.OptimizationService,
) *AggregatorHandler {
	return &AggregatorHandler{
		aggregatorService:   aggregatorService,
		metricsService:      metricsService,
		trainingService:     trainingService,
		optimizationService: optimizationService,
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

	aggregators, err := h.aggregatorService.GetAggregatorsByUser(userID)
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
	aggregator, err := h.aggregatorService.GetAggregatorByID(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	c.JSON(http.StatusOK, aggregator)
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

	// 요청 데이터 검증
	if err := aggregatorvalidator.ValidateCreateAggregatorRequest(
		request.Name,
		request.Algorithm,
		request.Region,
		request.Storage,
		request.InstanceType,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 서비스 입력 구성
	input := aggregator.CreateAggregatorInput{
		Name:         request.Name,
		Algorithm:    request.Algorithm,
		Region:       request.Region,
		Storage:      request.Storage,
		InstanceType: request.InstanceType,
		UserID:       userID,
	}

	// Aggregator 생성
	result, err := h.aggregatorService.CreateAggregator(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 생성 실패: " + err.Error()})
		return
	}

	// 응답 반환
	response := CreateAggregatorResponse{
		AggregatorID:    result.AggregatorID,
		Status:          result.Status,
		TerraformStatus: result.TerraformStatus,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
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
	
	if err := h.aggregatorService.DeleteAggregator(id, userID); err != nil {
		if err == aggregator.ErrAggregatorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 삭제 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aggregator가 삭제되었습니다"})
}