package aggregator

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/Mungge/Fleecy-Cloud/validators/aggregator"
)

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
	
	var request UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 요청 데이터 검증
	if err := aggregatorvalidator.ValidateUpdateStatusRequest(request.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.metricsService.UpdateStatus(id, userID, request.Status); err != nil {
		if err == aggregator.ErrAggregatorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
			return
		}
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
	
	var request UpdateMetricsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 요청 데이터 검증
	if err := aggregatorvalidator.ValidateUpdateMetricsRequest(request.CPUUsage, request.MemoryUsage, request.NetworkUsage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.metricsService.UpdateMetrics(id, userID, request.CPUUsage, request.MemoryUsage, request.NetworkUsage)
	if err != nil {
		if err == aggregator.ErrAggregatorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "메트릭 업데이트 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "메트릭이 업데이트되었습니다"})
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

	stats, err := h.metricsService.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, stats)
}