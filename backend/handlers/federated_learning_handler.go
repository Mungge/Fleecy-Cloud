package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// FederatedLearningHandler는 연합학습 관련 API 핸들러입니다
type FederatedLearningHandler struct {
	repo *repository.FederatedLearningRepository
}

// NewFederatedLearningHandler는 새 FederatedLearningHandler 인스턴스를 생성합니다
func NewFederatedLearningHandler(repo *repository.FederatedLearningRepository) *FederatedLearningHandler {
	return &FederatedLearningHandler{repo: repo}
}


// GetFederatedLearnings는 사용자의 모든 연합학습 작업을 반환하는 핸들러입니다
func (h *FederatedLearningHandler) GetFederatedLearnings(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	// 사용자의 모든 연합학습 작업 조회
	fls, err := h.repo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fls})
}

// GetFederatedLearning은 특정 ID의 연합학습 작업을 반환하는 핸들러입니다
func (h *FederatedLearningHandler) GetFederatedLearning(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 연합학습 ID 추출
	id := c.Param("id")

	// DB에서 연합학습 조회
	fl, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 조회에 실패했습니다"})
		return
	}
	if fl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "연합학습 작업을 찾을 수 없습니다"})
		return
	}

	// 작업 소유자 확인
	if fl.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 연합학습 작업에 접근할 권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fl})
}

// UpdateFederatedLearning은 연합학습 작업을 업데이트하는 핸들러입니다
func (h *FederatedLearningHandler) UpdateFederatedLearning(c *gin.Context) {
	// 사용자 ID 추출
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	// 경로 매개변수에서 연합학습 ID 추출
	id := c.Param("id")

	// DB에서 연합학습 조회
	fl, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 조회에 실패했습니다"})
		return
	}
	if fl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "연합학습 작업을 찾을 수 없습니다"})
		return
	}

	// 작업 소유자 확인
	if fl.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 연합학습 작업을 수정할 권한이 없습니다"})
		return
	}

	// 요청 본문 파싱
	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		ModelType   string   `json:"modelType"`
		Algorithm   string   `json:"algorithm"`
		Rounds      int      `json:"rounds"`
		Participants []string `json:"participants"`
		Accuracy    string   `json:"accuracy"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 필드 업데이트
	if request.Name != "" {
		fl.Name = request.Name
	}
	if request.Description != "" {
		fl.Description = request.Description
	}
	if request.Status != "" {
		fl.Status = request.Status
		
		// 작업이 완료된 경우 완료 시간 설정
		if request.Status == "완료" {
			now := time.Now()
			fl.CompletedAt = &now
		}
	}
	if request.ModelType != "" {
		fl.ModelType = request.ModelType
	}
	if request.Algorithm != "" {
		fl.Algorithm = request.Algorithm
	}
	if request.Rounds > 0 {
		fl.Rounds = request.Rounds
	}
	if len(request.Participants) > 0 {
		fl.ParticipantCount = len(request.Participants)
	}
	if request.Accuracy != "" {
		fl.Accuracy = request.Accuracy
	}

	fl.UpdatedAt = time.Now()

	// DB 업데이트
	if err := h.repo.Update(fl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 업데이트에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fl})
}

// DeleteFederatedLearning은 연합학습 작업을 삭제
func (h *FederatedLearningHandler) DeleteFederatedLearning(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 연합학습 ID 추출
	id := c.Param("id")

	// DB에서 연합학습 조회
	fl, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 조회에 실패했습니다"})
		return
	}
	if fl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "연합학습 작업을 찾을 수 없습니다"})
		return
	}

	// 작업 소유자 확인
	if fl.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 연합학습 작업을 삭제할 권한이 없습니다"})
		return
	}

	// DB에서 삭제
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 삭제에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "연합학습 작업이 삭제되었습니다"})
}

// CreateFederatedLearningForAggregator godoc
// @Summary 기존 Aggregator에 연합학습 생성
// @Description 기존 Aggregator ID를 받아서 연합학습을 생성합니다.
// @Tags federated-learning
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Param federatedLearning body CreateFederatedLearningForAggregatorRequest true "연합학습 생성 정보"
// @Success 201 {object} CreateFederatedLearningResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/federated-learning [post]
func (h *FederatedLearningHandler) CreateFederatedLearning(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")
	if aggregatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aggregator ID가 필요합니다"})
		return
	}

	var request CreateFederatedLearningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다: " + err.Error()})
		return
	}

	if request.CloudConnectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CloudConnectionID가 필요합니다"})
		return
	}

	// FederatedLearning 생성
	federatedLearning := &models.FederatedLearning{
		ID:                uuid.New().String(),
		UserID:           userID,
		CloudConnectionID: request.CloudConnectionID,
		AggregatorID:     &aggregatorID,
		Name:             request.Name,
		Description:      request.Description,
		Status:           "ready",
		ParticipantCount: len(request.Participants),
		Rounds:           request.Rounds,
		Algorithm:        request.Algorithm,
		ModelType:        request.ModelType,
	}

	// DB에 저장
	if err := h.repo.Create(federatedLearning); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 생성 실패: " + err.Error()})
		return
	}

	// 응답 반환
	response := CreateFederatedLearningResponse{
		FederatedLearningID: federatedLearning.ID,
		AggregatorID:        aggregatorID,
		Status:              "ready",
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// FederatedLearning 생성 요청 구조 (AggregatorID 기반)
type CreateFederatedLearningRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	ModelType    string `json:"modelType" binding:"required"`
	Algorithm    string `json:"algorithm" binding:"required"`
	Rounds       int    `json:"rounds" binding:"required"`
	Participants []struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		Status            string `json:"status"`
		OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
	} `json:"participants" binding:"required"`
	ModelFileName string `json:"modelFileName,omitempty"`
}

// FederatedLearning 생성 응답 구조
type CreateFederatedLearningResponse struct {
	FederatedLearningID string `json:"federatedLearningId"`
	AggregatorID        string `json:"aggregatorId"`
	Status              string `json:"status"`
}