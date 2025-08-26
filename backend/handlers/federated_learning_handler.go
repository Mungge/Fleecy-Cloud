package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// FederatedLearningHandler는 연합학습 관련 API 핸들러입니다
type FederatedLearningHandler struct {
	repo            *repository.FederatedLearningRepository
	participantRepo *repository.ParticipantRepository
	aggregatorRepo  *repository.AggregatorRepository
}

// NewFederatedLearningHandler는 새 FederatedLearningHandler 인스턴스를 생성합니다
func NewFederatedLearningHandler(repo *repository.FederatedLearningRepository, participantRepo *repository.ParticipantRepository, aggregatorRepo *repository.AggregatorRepository) *FederatedLearningHandler {
	return &FederatedLearningHandler{
		repo:            repo,
		participantRepo: participantRepo,
		aggregatorRepo:  aggregatorRepo,
	}
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
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Status       string   `json:"status"`
		ModelType    string   `json:"modelType"`
		Algorithm    string   `json:"algorithm"`
		Rounds       int      `json:"rounds"`
		Participants []string `json:"participants"`
		Accuracy     string   `json:"accuracy"`
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

// CreateFederatedLearning godoc
// @Summary 연합학습 생성
// @Description Aggregator ID를 포함한 연합학습을 생성합니다.
// @Tags federated-learning
// @Accept json
// @Produce json
// @Param federatedLearning body CreateFederatedLearningRequest true "연합학습 생성 정보"
// @Success 201 {object} CreateFederatedLearningResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/federated-learning [post]
func (h *FederatedLearningHandler) CreateFederatedLearning(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	var request CreateFederatedLearningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다: " + err.Error()})
		return
	}

	if request.AggregatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aggregator ID가 필요합니다"})
		return
	}

	if request.CloudConnectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CloudConnectionID가 필요합니다"})
		return
	}

	// FederatedLearning 생성
	federatedLearning := &models.FederatedLearning{
		ID:                uuid.New().String(),
		UserID:            userID,
		CloudConnectionID: request.CloudConnectionID,
		AggregatorID:      &request.AggregatorID,
		Name:              request.Name,
		Description:       request.Description,
		Status:            "ready",
		ParticipantCount:  len(request.Participants),
		Rounds:            request.Rounds,
		Algorithm:         request.Algorithm,
		ModelType:         request.ModelType,
	}

	// 참여자 ID 추출
	var participantIDs []string
	for _, p := range request.Participants {
		participantIDs = append(participantIDs, p.ID)
	}

	// DB에 참여자와 함께 저장
	if err := h.repo.CreateWithParticipants(federatedLearning, participantIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 생성 실패: " + err.Error()})
		return
	}

	// 응답 반환
	response := CreateFederatedLearningResponse{
		FederatedLearningID: federatedLearning.ID,
		AggregatorID:        request.AggregatorID,
		Status:              "ready",
	}

	// 참여자들과 집계자에게 연합학습 실행 요청 전송
	go h.sendFederatedLearningExecuteRequests(federatedLearning, request.Participants)

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// sendFederatedLearningExecuteRequests는 집계자와 모든 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendFederatedLearningExecuteRequests(federatedLearning *models.FederatedLearning, participants []struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
}) {
	// 1. 먼저 집계자에게 실행 요청 전송
	if err := h.sendExecuteRequestToAggregatorFirst(federatedLearning); err != nil {
		fmt.Printf("집계자 실행 요청 실패: %v\n", err)
		return // 집계자 실행 요청이 실패하면 전체 프로세스 중단
	}

	// 2. 집계자 실행 요청이 성공한 후 참여자들에게 실행 요청 전송
	h.sendExecuteRequestToParticipants(federatedLearning, participants)
}

// sendExecuteRequestToParticipant는 개별 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToParticipant(participant *models.Participant, federatedLearning *models.FederatedLearning) error {
	// OpenStack 엔드포인트에서 포트 5000으로 요청 URL 구성
	endpoint := strings.TrimSuffix(participant.OpenStackEndpoint, "/")
	if strings.HasSuffix(endpoint, ":5000") {
		// 이미 5000 포트가 있는 경우
	} else if strings.Contains(endpoint, ":") {
		// 다른 포트가 있는 경우 5000으로 변경
		parts := strings.Split(endpoint, ":")
		endpoint = parts[0] + ":" + parts[1] + ":5000"
	} else {
		// 포트가 없는 경우 5000 추가
		endpoint = endpoint + ":5000"
	}

	requestURL := endpoint + "/api/fl/execute"

	// 요청 페이로드 구성
	payload := map[string]interface{}{
		"federated_learning_id": federatedLearning.ID,
		"participant_id":        participant.ID,
		"rounds":                federatedLearning.Rounds,
		"algorithm":             federatedLearning.Algorithm,
		"model_type":            federatedLearning.ModelType,
		"name":                  federatedLearning.Name,
		"description":           federatedLearning.Description,
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %v", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")

	// HTTP 클라이언트 생성 및 요청 전송
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 요청 전송 실패: %v", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP 요청 실패: 상태 코드 %d", resp.StatusCode)
	}

	return nil
}

// sendExecuteRequestToAggregatorFirst는 집계자에게 먼저 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToAggregatorFirst(federatedLearning *models.FederatedLearning) error {
	if federatedLearning.AggregatorID == nil {
		return fmt.Errorf("집계자 ID가 설정되지 않았습니다")
	}

	aggregatorData, err := h.aggregatorRepo.GetAggregatorByID(*federatedLearning.AggregatorID)
	if err != nil {
		return fmt.Errorf("집계자 조회 실패 (ID: %s): %v", *federatedLearning.AggregatorID, err)
	}

	if aggregatorData == nil {
		return fmt.Errorf("집계자를 찾을 수 없습니다 (ID: %s)", *federatedLearning.AggregatorID)
	}

	// 집계자 Public IP가 없으면 실패
	if aggregatorData.PublicIP == "" {
		return fmt.Errorf("집계자 %s의 Public IP가 설정되지 않았습니다", aggregatorData.Name)
	}

	// 집계자에게 연합학습 실행 요청 전송
	if err := h.sendExecuteRequestToAggregator(aggregatorData, federatedLearning); err != nil {
		return fmt.Errorf("집계자 %s에게 연합학습 실행 요청 전송 실패: %v", aggregatorData.Name, err)
	}

	fmt.Printf("집계자 %s에게 연합학습 실행 요청 전송 성공\n", aggregatorData.Name)
	return nil
}

// sendExecuteRequestToParticipants는 모든 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToParticipants(federatedLearning *models.FederatedLearning, participants []struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
}) {
	fmt.Printf("참여자 %d명에게 연합학습 실행 요청을 전송합니다\n", len(participants))

	for _, participant := range participants {
		// 참여자 정보 조회
		participantData, err := h.participantRepo.GetByID(participant.ID)
		if err != nil {
			fmt.Printf("참여자 조회 실패 (ID: %s): %v\n", participant.ID, err)
			continue
		}

		// OpenStack 엔드포인트가 없으면 스킵
		if participantData.OpenStackEndpoint == "" {
			fmt.Printf("참여자 %s의 OpenStack 엔드포인트가 설정되지 않았습니다\n", participantData.Name)
			continue
		}

		// 연합학습 실행 요청 전송
		if err := h.sendExecuteRequestToParticipant(participantData, federatedLearning); err != nil {
			fmt.Printf("참여자 %s에게 연합학습 실행 요청 전송 실패: %v\n", participantData.Name, err)
		} else {
			fmt.Printf("참여자 %s에게 연합학습 실행 요청 전송 성공\n", participantData.Name)
		}
	}
}

// sendExecuteRequestToAggregator는 집계자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToAggregator(aggregator *models.Aggregator, federatedLearning *models.FederatedLearning) error {
	// 집계자 Public IP에서 포트 5000으로 요청 URL 구성
	endpoint := "http://" + aggregator.PublicIP + ":5000"
	requestURL := endpoint + "/api/fl/execute"

	// 요청 페이로드 구성
	payload := map[string]interface{}{
		"federated_learning_id": federatedLearning.ID,
		"aggregator_id":         aggregator.ID,
		"rounds":                federatedLearning.Rounds,
		"algorithm":             federatedLearning.Algorithm,
		"model_type":            federatedLearning.ModelType,
		"name":                  federatedLearning.Name,
		"description":           federatedLearning.Description,
		"role":                  "aggregator", // 집계자 역할 명시
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %v", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")

	// HTTP 클라이언트 생성 및 요청 전송
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 요청 전송 실패: %v", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP 요청 실패: 상태 코드 %d", resp.StatusCode)
	}

	return nil
}

// FederatedLearning 생성 요청 구조 (AggregatorID 기반)
type CreateFederatedLearningRequest struct {
	AggregatorID      string `json:"aggregatorId" binding:"required"`
	CloudConnectionID string `json:"cloudConnectionId" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	ModelType         string `json:"modelType" binding:"required"`
	Algorithm         string `json:"algorithm" binding:"required"`
	Rounds            int    `json:"rounds" binding:"required"`
	Participants      []struct {
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
