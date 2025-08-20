package aggregator

import (
	"net/http"
	"time"

	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/gin-gonic/gin"
)

// DeploymentProgress 배포 진행 상황
type DeploymentProgress struct {
	AggregatorID string `json:"aggregator_id"`
	Stage        int    `json:"stage"`
	TotalStages  int    `json:"total_stages"`
	Message      string `json:"message"`
	Status       string `json:"status"` // "creating", "progress", "completed", "failed"
	Error        string `json:"error,omitempty"`
	LastUpdated  int64  `json:"last_updated"`
}

// GetAggregatorProgress godoc
// @Summary 집계자 배포 진행 상황 조회
// @Description 집계자 배포의 실시간 진행 상황을 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {object} DeploymentProgress
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/aggregators/{id}/progress [get]
func (h *AggregatorHandler) GetAggregatorProgress(c *gin.Context) {
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

	// 진행 상황 조회 (메모리 캐시 또는 DB에서)
	progress := getDeploymentProgress(aggregatorID, aggregator.Status)

	c.JSON(http.StatusOK, progress)
}

// getDeploymentProgress 배포 진행 상황 반환 (실제 구현에서는 캐시나 DB 사용)
func getDeploymentProgress(aggregatorID, status string) DeploymentProgress {
	switch status {
	case "creating":
		return DeploymentProgress{
			AggregatorID: aggregatorID,
			Stage:        1,
			TotalStages:  5,
			Message:      "배포 준비 중...",
			Status:       "progress",
			LastUpdated:  getCurrentTimestamp(),
		}
	case "running":
		return DeploymentProgress{
			AggregatorID: aggregatorID,
			Stage:        5,
			TotalStages:  5,
			Message:      "배포 완료",
			Status:       "completed",
			LastUpdated:  getCurrentTimestamp(),
		}
	case "failed":
		return DeploymentProgress{
			AggregatorID: aggregatorID,
			Stage:        0,
			TotalStages:  5,
			Message:      "배포 실패",
			Status:       "failed",
			Error:        "배포 중 오류가 발생했습니다",
			LastUpdated:  getCurrentTimestamp(),
		}
	default:
		return DeploymentProgress{
			AggregatorID: aggregatorID,
			Stage:        0,
			TotalStages:  5,
			Message:      "상태 불명",
			Status:       "unknown",
			LastUpdated:  getCurrentTimestamp(),
		}
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix() * 1000
}
