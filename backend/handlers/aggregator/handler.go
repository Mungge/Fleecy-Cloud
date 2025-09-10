package aggregator

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/Mungge/Fleecy-Cloud/utils"
	aggregatorvalidator "github.com/Mungge/Fleecy-Cloud/validators/aggregator"
)

// sanitizeGCPName은 GCP 리소스 이름 규칙에 맞게 이름을 변환합니다
// GCP 규칙: 소문자, 숫자, 하이픈만 허용, 1-63자, 시작과 끝은 문자나 숫자
func sanitizeGCPName(name string) string {
	// 소문자로 변환
	name = strings.ToLower(name)
	
	// 허용되지 않는 문자를 하이픈으로 변환
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "-")
	
	// 연속된 하이픈을 하나로 변환
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")
	
	// 시작과 끝의 하이픈 제거
	name = strings.Trim(name, "-")
	
	// 길이 제한 (63자)
	if len(name) > 63 {
		name = name[:63]
		name = strings.Trim(name, "-")
	}
	
	// 빈 문자열이면 기본값 사용
	if name == "" {
		name = "fleecy-project"
	}
	
	return name
}

// sanitizeGCPServiceAccountID는 GCP Service Account ID 규칙에 맞게 이름을 변환합니다
// GCP 규칙: ^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$ (6-30자, 시작과 끝이 문자)
func sanitizeGCPServiceAccountID(name string) string {
	// 소문자로 변환
	name = strings.ToLower(name)
	
	// 허용되지 않는 문자를 하이픈으로 변환
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "-")
	
	// 연속된 하이픈을 하나로 변환
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")
	
	// 시작과 끝의 하이픈 제거
	name = strings.Trim(name, "-")
	
	// 길이 제한 (30자)
	if len(name) > 30 {
		name = name[:30]
		name = strings.Trim(name, "-")
	}
	
	// 최소 길이 보장 (6자)
	if len(name) < 6 {
		name = name + "sa"
	}
	
	// 시작이 문자인지 확인
	if len(name) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(name) {
		name = "f" + name
	}
	
	// 끝이 문자나 숫자인지 확인
	if len(name) > 0 && !regexp.MustCompile(`[a-z0-9]$`).MatchString(name) {
		name = name + "1"
	}
	
	// 빈 문자열이면 기본값 사용
	if name == "" {
		name = "fleecy-sa"
	}
	
	return name
}

// getZoneName은 클라우드 프로바이더에 따라 올바른 존/가용영역 이름을 반환합니다
func getZoneName(cloudProvider, region string) string {
	switch strings.ToLower(cloudProvider) {
	case "aws":
		// AWS 형식: us-east-1a, us-east-1b, us-east-1c
		return region + "a"
	case "gcp":
		// GCP 형식: us-central1-a, us-central1-b, us-central1-c
		return region + "-a"
	default:
		// 기본값 (AWS 형식)
		return region + "a"
	}
}

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
		request.EstimatedCost,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := aggregator.CreateAggregatorInput{
		Name:          request.Name,
		Algorithm:     request.Algorithm,
		Region:        request.Region,
		Storage:       request.Storage,
		InstanceType:  request.InstanceType,
		UserID:        userID,
		CloudProvider: request.CloudProvider,
		ProjectName:   sanitizeGCPName(request.Name + "-project"),
		Zone:          getZoneName(request.CloudProvider, request.Region),
		EstimatedCost: request.EstimatedCost,
	}

	// 타임아웃 설정 (최대 10분)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	// Aggregator 생성 및 배포 (동기 처리)
	result, err := h.aggregatorService.CreateAggregatorWithContext(ctx, input)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Aggregator 배포가 타임아웃되었습니다. 나중에 상태를 확인해주세요.",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Aggregator 생성 및 배포 실패: " + err.Error(),
			})
		}
		return
	}

	// 응답 반환 (배포 완료된 상태)
	response := CreateAggregatorResponse{
		AggregatorID:    result.AggregatorID,
		Status:          result.Status,          // "running" 또는 "failed"
		TerraformStatus: result.TerraformStatus, // "completed"
	}

	// 성공 시 201, 배포 완료 메시지와 함께 반환
	c.JSON(http.StatusCreated, gin.H{
		"message": "Aggregator가 성공적으로 생성되고 배포되었습니다",
		"data":    response,
	})
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

// getStringValue 헬퍼 함수: *string을 string으로 변환
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// WebSocketProgress godoc
// @Summary 집계자 배포 진행 상황 WebSocket 연결
// @Description 집계자 배포의 실시간 진행 상황을 WebSocket으로 전송합니다.
// @Tags aggregators
// @Param id path string true "Aggregator ID"
// @Router /api/aggregators/{id}/progress/ws [get]
func (h *AggregatorHandler) WebSocketProgress(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")

	// 권한 확인
	aggregator, err := h.aggregatorService.GetAggregatorByID(aggregatorID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// WebSocket 연결 업그레이드
	h.aggregatorService.HandleWebSocketProgress(c.Writer, c.Request, aggregatorID)
}
