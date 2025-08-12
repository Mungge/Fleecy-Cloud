package aggregator

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

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
	
	rounds, err := h.trainingService.GetTrainingHistory(id, userID)
	if err != nil {
		if err == aggregator.ErrAggregatorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "학습 히스토리 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, rounds)
}