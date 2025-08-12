package aggregator

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/Mungge/Fleecy-Cloud/validators/aggregator"
)

// OptimizeAggregatorPlacement godoc
// @Summary 집계자 배치 최적화
// @Description NSGA-II 알고리즘을 사용하여 사용자 제약사항에 맞는 최적의 집계자 배치 위치를 찾습니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param request body aggregator.OptimizationRequest true "최적화 요청 정보"
// @Success 200 {object} map[string]interface{} "data: aggregator.OptimizationResponse"
// @Failure 400 {object} map[string]string "error: 잘못된 요청"
// @Failure 401 {object} map[string]string "error: 인증 필요"
// @Failure 500 {object} map[string]string "error: 서버 오류"
// @Router /api/aggregators/optimization [post]
func (h *AggregatorHandler) OptimizeAggregatorPlacement(c *gin.Context) {
	// 1. 사용자 인증 확인
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	// 2. 요청 데이터 파싱
	var request aggregator.OptimizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "잘못된 요청 형식입니다: " + err.Error(),
		})
		return
	}

	// 3. 요청 데이터 유효성 검증
	if err := aggregatorvalidator.ValidateOptimizationRequest(&request); err != nil {
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