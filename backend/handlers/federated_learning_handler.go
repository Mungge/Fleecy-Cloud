package handlers

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

//go:embed templates/server_app.py
var serverAppTemplate string

//go:embed templates/client_app.py
var clientAppTemplate string

//go:embed templates/task.py
var taskTemplate string

//go:embed templates/pyproject.toml
var pyprojectTemplate string

// FederatedLearningHandler는 연합학습 관련 API 핸들러입니다
type FederatedLearningHandler struct {
	repo              *repository.FederatedLearningRepository
	participantRepo   *repository.ParticipantRepository
	aggregatorRepo    *repository.AggregatorRepository
	sshKeypairService *services.SSHKeypairService
}

// NewFederatedLearningHandler는 새 FederatedLearningHandler 인스턴스를 생성합니다
func NewFederatedLearningHandler(repo *repository.FederatedLearningRepository, participantRepo *repository.ParticipantRepository, aggregatorRepo *repository.AggregatorRepository, sshKeypairService *services.SSHKeypairService) *FederatedLearningHandler {
	return &FederatedLearningHandler{
		repo:              repo,
		participantRepo:   participantRepo,
		aggregatorRepo:    aggregatorRepo,
		sshKeypairService: sshKeypairService,
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

	// MLflow URL 생성 (집계자가 설정된 경우)
	var mlflowURL string
	if fl.AggregatorID != nil {
		aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
		if err == nil && aggregator != nil {
			mlflowURL = fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
		}
	}

	response := gin.H{
		"federatedLearning": fl,
		"mlflowURL":         mlflowURL,
		"experimentName":    fmt.Sprintf("federated-learning-%s", fl.ID),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
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

	// 포트 9092 고정

	// 참여자들과 집계자에게 연합학습 실행 요청 전송
	fmt.Printf("🚀 고루틴 시작: 연합학습 실행 요청 전송 프로세스 시작 (ID: %s)\n", federatedLearning.ID)
	go func() {
		fmt.Printf("📍 고루틴 내부: sendFederatedLearningExecuteRequests 함수 호출 시작\n")
		h.sendFederatedLearningExecuteRequests(federatedLearning, request.Participants)
		fmt.Printf("✅ 고루틴 완료: sendFederatedLearningExecuteRequests 함수 호출 완료\n")
	}()

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// sendFederatedLearningExecuteRequests는 집계자와 모든 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendFederatedLearningExecuteRequests(federatedLearning *models.FederatedLearning, participants []struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
}) {
	fmt.Printf("🔄 연합학습 실행 요청 프로세스 시작 - FL ID: %s\n", federatedLearning.ID)
	fmt.Printf("📊 전달받은 참여자 수: %d명\n", len(participants))

	// 1. 먼저 집계자에게 실행 요청 전송
	fmt.Printf("1️⃣ 단계 1: 집계자 실행 요청 전송\n")
	if federatedLearning.AggregatorID == nil {
		fmt.Printf("❌ 집계자 ID가 설정되지 않았습니다\n")
		return
	}

	fmt.Printf("🔍 집계자 조회 중 - ID: %s\n", *federatedLearning.AggregatorID)
	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*federatedLearning.AggregatorID)
	if err != nil {
		fmt.Printf("❌ 집계자 조회 실패: %v\n", err)
		return
	}

	if aggregator == nil {
		fmt.Printf("❌ 집계자를 찾을 수 없습니다\n")
		return
	}

	fmt.Printf("✅ 집계자 조회 성공 - Name: %s, IP: %s\n", aggregator.Name, aggregator.PublicIP)
	fmt.Printf("🚀 집계자에게 실행 요청 전송 중...\n")
	if err := h.sendExecuteRequestToAggregator(aggregator, federatedLearning); err != nil {
		fmt.Printf("❌ 집계자 실행 요청 실패: %v\n", err)
		return // 집계자 실행 요청이 실패하면 전체 프로세스 중단
	}
	fmt.Printf("✅ 집계자 실행 요청 성공\n")

	// 2. 집계자 실행 요청이 성공한 후, 집계자 서버가 준비될 때까지 대기
	fmt.Printf("2️⃣ 단계 2: 집계자 서버 준비 상태 확인 중...\n")
	if err := h.waitForAggregatorReady(aggregator); err != nil {
		fmt.Printf("❌ 집계자 서버 준비 대기 실패: %v\n", err)
		return
	}

	fmt.Printf("✅ 집계자 서버 준비 완료!\n")
	fmt.Printf("3️⃣ 단계 3: 참여자들에게 요청을 전송합니다.\n")
	// 3. 집계자가 준비된 후 참여자들에게 실행 요청 전송
	h.sendExecuteRequestToParticipants(federatedLearning, participants)
	fmt.Printf("🎉 연합학습 실행 요청 프로세스 완료\n")
}

// sendExecuteRequestToParticipant는 개별 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToParticipant(participant *models.Participant, federatedLearning *models.FederatedLearning) error {

	requestURL := fmt.Sprintf("%s:5000/api/fl/execute-local", participant.OpenStackEndpoint)

	fmt.Printf("참여자 서버 URL: %s\n", requestURL)

	// 집계자 주소 가져오기
	aggregatorAddress, err := h.getAggregatorAddress(federatedLearning)
	if err != nil {
		return fmt.Errorf("집계자 주소 조회 실패: %v", err)
	}

	// 새로운 로컬 실행 API를 위한 페이로드 구성
	payload := map[string]interface{}{
		"server_address": aggregatorAddress,
		"local_epochs":   5,   // 기본값 5로 설정 (COVID-19 데이터셋에 적합)
		"timeout":        600, // 10분 타임아웃
		"files": map[string]interface{}{
			"client_app.py": clientAppTemplate,
			"task.py":       taskTemplate,
		},
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %v", err)
	}

	// 요청 로깅
	fmt.Printf("=== 참여자 %s에게 로컬 실행 요청 전송 ===\n", participant.ID)
	fmt.Printf("요청 URL: %s\n", requestURL)
	fmt.Printf("집계자 주소: %s\n", aggregatorAddress)
	fmt.Printf("요청 페이로드: %s\n", string(jsonData))

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")

	// HTTP 클라이언트 생성 및 요청 전송 (패키지 설치 시간 고려하여 타임아웃 증가)
	client := &http.Client{
		Timeout: 120 * time.Second, // 2분 타임아웃 (패키지 설치 + 초기 응답)
	}

	fmt.Printf("HTTP 요청 전송 중...\n")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 참여자 %s에게 요청 전송 실패: %v\n", participant.ID, err)
		return fmt.Errorf("HTTP 요청 전송 실패: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("응답 상태 코드: %d\n", resp.StatusCode)

	// 응답 본문 읽기 (디버깅용)
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("응답 본문 읽기 실패: %v\n", err)
		responseBody = []byte("응답 본문 읽기 실패")
	}

	// 응답 상태 코드 확인
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("❌ 참여자 %s 요청 실패: 상태 코드 %d\n", participant.ID, resp.StatusCode)
		fmt.Printf("응답 본문: %s\n", string(responseBody))
		return fmt.Errorf("HTTP 요청 실패: 상태 코드 %d, 응답: %s", resp.StatusCode, string(responseBody))
	}

	fmt.Printf("✅ 참여자 %s에게 로컬 실행 요청 전송 성공 (상태 코드: %d)\n", participant.ID, resp.StatusCode)
	fmt.Printf("응답 본문: %s\n", string(responseBody))
	fmt.Printf("=== 요청 완료 ===\n\n")
	return nil
}

// sendExecuteRequestToParticipants는 모든 참여자에게 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToParticipants(federatedLearning *models.FederatedLearning, participants []struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	OpenstackEndpoint string `json:"openstack_endpoint,omitempty"`
}) {
	fmt.Printf("🎯 참여자 요청 전송 시작: %d명의 참여자에게 연합학습 실행 요청을 전송합니다\n", len(participants))

	if len(participants) == 0 {
		fmt.Printf("⚠️ 경고: 참여자가 없습니다. 요청을 전송할 대상이 없습니다.\n")
		return
	}

	for i, participant := range participants {
		fmt.Printf("📋 참여자 %d/%d 처리 중 - ID: %s, Name: %s\n", i+1, len(participants), participant.ID, participant.Name)

		// 참여자 정보 조회
		participantData, err := h.participantRepo.GetByID(participant.ID)
		if err != nil {
			fmt.Printf("❌ 참여자 조회 실패 (ID: %s): %v\n", participant.ID, err)
			continue
		}

		if participantData == nil {
			fmt.Printf("❌ 참여자 데이터가 nil입니다 (ID: %s)\n", participant.ID)
			continue
		}

		fmt.Printf("✅ 참여자 조회 성공 - Name: %s, Endpoint: %s\n", participantData.Name, participantData.OpenStackEndpoint)

		// OpenStack 엔드포인트가 없으면 스킵
		if participantData.OpenStackEndpoint == "" {
			fmt.Printf("⚠️ 참여자 %s의 엔드포인트가 설정되지 않았습니다. 스킵합니다.\n", participantData.Name)
			continue
		}

		// 연합학습 실행 요청 전송
		fmt.Printf("🚀 참여자 %s에게 연합학습 실행 요청 전송 시작...\n", participantData.Name)
		if err := h.sendExecuteRequestToParticipant(participantData, federatedLearning); err != nil {
			fmt.Printf("❌ 참여자 %s에게 연합학습 실행 요청 전송 실패: %v\n", participantData.Name, err)
		} else {
			fmt.Printf("✅ 참여자 %s에게 연합학습 실행 요청 전송 성공\n", participantData.Name)
		}
	}

	fmt.Printf("🏁 모든 참여자에 대한 요청 전송 프로세스 완료\n")
}

// sendExecuteRequestToAggregator는 집계자에게 SSH를 통해 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToAggregator(aggregator *models.Aggregator, federatedLearning *models.FederatedLearning) error {
	fmt.Printf("🔧 집계자 실행 요청 시작 - Name: %s, IP: %s\n", aggregator.Name, aggregator.PublicIP)

	// 집계자 Public IP 확인
	if aggregator.PublicIP == "" {
		fmt.Printf("❌ 집계자 Public IP가 설정되지 않음\n")
		return fmt.Errorf("집계자 %s의 Public IP가 설정되지 않았습니다", aggregator.Name)
	}
	fmt.Printf("✅ 집계자 Public IP 확인 완료: %s\n", aggregator.PublicIP)

	// SSH 키페어 조회
	fmt.Printf("🔑 SSH 키페어 조회 중...\n")
	keypairWithPrivateKey, err := h.sshKeypairService.GetKeypairWithPrivateKey(aggregator.ID)
	if err != nil {
		fmt.Printf("❌ SSH 키페어 조회 실패: %v\n", err)
		return fmt.Errorf("집계자 %s의 SSH 키페어 조회 실패: %v", aggregator.Name, err)
	}
	fmt.Printf("✅ SSH 키페어 조회 성공\n")

	// SSH 클라이언트 생성
	fmt.Printf("🔗 SSH 클라이언트 생성 중...\n")
	sshClient := utils.NewSSHClient(
		aggregator.PublicIP,
		"22",
		"ubuntu",
		keypairWithPrivateKey.PrivateKey,
	)
	fmt.Printf("✅ SSH 클라이언트 생성 완료\n")

	// SSH 연결 테스트
	fmt.Printf("🔌 SSH 연결 테스트 중... (IP: %s, 타임아웃: 30초)\n", aggregator.PublicIP)
	startTime := time.Now()
	if err := sshClient.CheckConnection(); err != nil {
		elapsed := time.Since(startTime)
		fmt.Printf("❌ SSH 연결 실패 (소요시간: %.2f초): %v\n", elapsed.Seconds(), err)
		return fmt.Errorf("집계자 %s SSH 연결 실패: %v", aggregator.Name, err)
	}
	elapsed := time.Since(startTime)
	fmt.Printf("✅ SSH 연결 성공 (소요시간: %.2f초)\n", elapsed.Seconds())

	// 작업 디렉토리 생성
	workDir := fmt.Sprintf("/home/ubuntu/fl-aggregator-%s", federatedLearning.ID)
	_, _, err = sshClient.ExecuteCommand(fmt.Sprintf("mkdir -p %s", workDir))
	if err != nil {
		return fmt.Errorf("작업 디렉토리 생성 실패: %v", err)
	}

	// 집계자 주소 가져오기
	aggregatorAddress, err := h.getAggregatorAddress(federatedLearning)
	if err != nil {
		return fmt.Errorf("집계자 주소 조회 실패: %v", err)
	}

	// pyproject.toml 파일을 동적으로 생성 (참여자 수, 라운드 수, 집계자 주소 반영)
	dynamicPyprojectContent := strings.ReplaceAll(pyprojectTemplate, "min-fit-clients = 1", fmt.Sprintf("min-fit-clients = %d", federatedLearning.ParticipantCount))
	dynamicPyprojectContent = strings.ReplaceAll(dynamicPyprojectContent, "min-available-clients = 1", fmt.Sprintf("min-available-clients = %d", federatedLearning.ParticipantCount))
	dynamicPyprojectContent = strings.ReplaceAll(dynamicPyprojectContent, "num-server-rounds = 10", fmt.Sprintf("num-server-rounds = %d", federatedLearning.Rounds))
	dynamicPyprojectContent = strings.ReplaceAll(dynamicPyprojectContent, "address = \"<HOST>:<PORT>\"", fmt.Sprintf("address = \"%s\"", aggregatorAddress))

	err = sshClient.UploadFileContent(dynamicPyprojectContent, fmt.Sprintf("%s/pyproject.toml", workDir))
	if err != nil {
		return fmt.Errorf("pyproject.toml 파일 업로드 실패: %v", err)
	}

	// Python 패키지용 __init__.py 파일 생성
	err = sshClient.UploadFileContent("", fmt.Sprintf("%s/__init__.py", workDir))
	if err != nil {
		return fmt.Errorf("__init__.py 파일 업로드 실패: %v", err)
	}

	// server_app.py 파일을 작업 디렉토리에 직접 업로드
	err = sshClient.UploadFileContent(serverAppTemplate, fmt.Sprintf("%s/server_app.py", workDir))
	if err != nil {
		return fmt.Errorf("서버 앱 파일 업로드 실패: %v", err)
	}

	// task.py 파일 업로드
	err = sshClient.UploadFileContent(taskTemplate, fmt.Sprintf("%s/task.py", workDir))
	if err != nil {
		return fmt.Errorf("task.py 파일 업로드 실패: %v", err)
	}

	// client_app.py 파일을 작업 디렉토리에 직접 업로드
	err = sshClient.UploadFileContent(clientAppTemplate, fmt.Sprintf("%s/client_app.py", workDir))
	if err != nil {
		return fmt.Errorf("클라이언트 앱 파일 업로드 실패: %v", err)
	}

	// Flower 서버와 MLflow 서버 실행 스크립트 생성
	runScript := `#!/bin/bash
echo "=== Flower 서버 및 MLflow 서버 설정 시작 ==="

# 시스템 업데이트
echo "시스템 패키지 업데이트 중..."
sudo apt update
sudo apt install -y python3-venv python3-pip

# 가상환경 생성 및 활성화
echo "Python 가상환경을 설정합니다..."
if [ ! -d "venv" ]; then
    echo "가상환경을 생성합니다..."
    python3 -m venv venv
fi

echo "가상환경을 활성화합니다..."
source venv/bin/activate

# 가상환경이 제대로 활성화되었는지 확인
echo "현재 Python 경로: $(which python)"
echo "현재 pip 경로: $(which pip)"

# pip 업그레이드
echo "pip를 업그레이드합니다..."
pip install --upgrade pip

# 필수 Python 패키지 설치 (MLflow 포함)
echo "필수 Python 패키지를 설치합니다..."
pip install uv
uv pip install flwr torch torchvision tomli scikit-learn mlflow

# 설치된 패키지 확인
echo "설치된 패키지 확인:"
pip list | grep -E "(flwr|torch|tomli|scikit-learn|mlflow)"

echo "Python 패키지 설치가 완료되었습니다."

# MLflow 서버 백그라운드 실행
echo "MLflow 서버를 시작합니다..."
export MLFLOW_TRACKING_URI="file:./mlruns"
export MLFLOW_EXPERIMENT_NAME="federated-learning-` + federatedLearning.ID + `"
nohup mlflow server --backend-store-uri file:./mlruns --default-artifact-root ./mlruns --host 0.0.0.0 --port 5000 > mlflow.log 2>&1 &
echo "MLflow 서버가 포트 5000에서 시작되었습니다."

# MLflow 서버 시작 대기
sleep 5

# Flower 서버 실행
echo "Flower 서버를 시작합니다..."
echo "참여자 수: ` + fmt.Sprintf("%d", federatedLearning.ParticipantCount) + `"
echo "라운드 수: ` + fmt.Sprintf("%d", federatedLearning.Rounds) + `"
echo "포트: 9092"

# 서버 준비 상태 파일 생성 함수
create_ready_file() {
    echo "서버가 준비되었습니다 - $(date)" > server_ready.txt
    echo "✅ 서버 준비 상태 파일 생성됨"
}

	// 서버 실행 후 준비 완료 표시
(
    source venv/bin/activate && \
    python3 server_app.py --server-address 0.0.0.0:9092 --num-rounds ` + fmt.Sprintf("%d", federatedLearning.Rounds) + ` --min-fit-clients ` + fmt.Sprintf("%d", federatedLearning.ParticipantCount) + ` --min-available-clients ` + fmt.Sprintf("%d", federatedLearning.ParticipantCount) + ` &
    
    # 서버 프로세스 시작 후 충분한 대기 시간
    echo "Flower 서버 프로세스 시작 대기 중..."
    sleep 10
    
    # 서버가 실제로 포트를 리스닝하고 있는지 확인 (더 긴 대기 시간)
    timeout=60  # 1분 대기
    count=0
    while ! nc -z localhost 9092 && [ $count -lt $timeout ]; do
        echo "Flower 서버 시작 대기 중... ($count/$timeout)"
        sleep 2
        count=$((count + 2))
    done
    
    if nc -z localhost 9092; then
        echo "✅ Flower 서버가 포트 9092에서 정상적으로 시작됨"
        create_ready_file
    else
        echo "❌ Flower 서버 시작 실패 - 포트 9092에서 응답 없음"
        exit 1
    fi
    
    # 백그라운드 프로세스 대기
    wait
) &echo "Flower 서버 시작 스크립트가 백그라운드에서 실행 중입니다."
`

	err = sshClient.UploadFileContent(runScript, fmt.Sprintf("%s/run_server.sh", workDir))
	if err != nil {
		return fmt.Errorf("실행 스크립트 업로드 실패: %v", err)
	}

	// 스크립트 실행 권한 부여
	_, _, err = sshClient.ExecuteCommand(fmt.Sprintf("chmod +x %s/run_server.sh", workDir))
	if err != nil {
		return fmt.Errorf("스크립트 권한 설정 실패: %v", err)
	}

	// 백그라운드에서 Flower 서버 실행
	command := fmt.Sprintf("cd %s && nohup ./run_server.sh > flower_server.log 2>&1 &", workDir)
	stdout, stderr, err := sshClient.ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("flower 서버 실행 실패: %v, stdout: %s, stderr: %s", err, stdout, stderr)
	}

	fmt.Printf("집계자 %s에서 Flower 서버 실행 성공\n", aggregator.Name)
	fmt.Printf("서버가 완전히 시작될 때까지 잠시 대기합니다...\n")

	// 서버 시작을 위한 초기 대기 시간 (30초)
	time.Sleep(30 * time.Second)

	return nil
}

// getAggregatorAddress는 집계자의 주소를 반환합니다 (포트 9092 고정)
func (h *FederatedLearningHandler) getAggregatorAddress(federatedLearning *models.FederatedLearning) (string, error) {
	if federatedLearning.AggregatorID == nil {
		return "", fmt.Errorf("집계자 ID가 설정되지 않았습니다")
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*federatedLearning.AggregatorID)
	if err != nil {
		return "", fmt.Errorf("집계자 조회 실패: %v", err)
	}

	if aggregator == nil {
		return "", fmt.Errorf("집계자를 찾을 수 없습니다")
	}

	if aggregator.PublicIP == "" {
		return "", fmt.Errorf("집계자의 Public IP가 설정되지 않았습니다")
	}

	// 포트 9092 고정
	return fmt.Sprintf("%s:9092", aggregator.PublicIP), nil
}

// waitForAggregatorReady는 집계자 서버가 준비될 때까지 대기합니다
func (h *FederatedLearningHandler) waitForAggregatorReady(aggregator *models.Aggregator) error {
	maxRetries := 40                  // 최대 40번 시도 (약 7분)
	retryInterval := 10 * time.Second // 10초 간격

	fmt.Printf("집계자 서버 준비 상태 확인 중... (로그 파일 기반)\n")
	fmt.Printf("집계자 IP: %s\n", aggregator.PublicIP)

	// SSH 키페어 조회
	fmt.Printf("🔑 대기 중 SSH 키페어 조회...\n")
	keypairWithPrivateKey, err := h.sshKeypairService.GetKeypairWithPrivateKey(aggregator.ID)
	if err != nil {
		return fmt.Errorf("SSH 키페어 조회 실패: %v", err)
	}
	fmt.Printf("✅ SSH 키페어 조회 성공\n")

	for i := 0; i < maxRetries; i++ {
		sshClient := utils.NewSSHClient(aggregator.PublicIP, "22", "ubuntu", keypairWithPrivateKey.PrivateKey)
		if sshClient != nil {
			// 1. 로그 파일에서 Flower 서버 시작 확인
			fmt.Printf("📋 로그 파일 확인 중... (시도 %d/%d)\n", i+1, maxRetries)
			logOutput, logStderr, logErr := sshClient.ExecuteCommand("tail -20 ~/federated_learning/flower_server.log")

			if logErr != nil {
				fmt.Printf("⚠️ 로그 파일 읽기 실패: %v, stderr: %s\n", logErr, logStderr)
			} else if len(strings.TrimSpace(logOutput)) == 0 {
				fmt.Printf("⚠️ 로그 파일이 비어있음\n")
			} else {
				fmt.Printf("📋 현재 로그 내용 (마지막 20줄):\n%s\n", logOutput)

				// Flower 서버가 시작되었는지 확인하는 키워드들
				if strings.Contains(logOutput, "Starting Flower server") &&
					strings.Contains(logOutput, "gRPC server running") {
					fmt.Printf("✅ 집계자 서버 준비 완료! Flower 서버 시작 로그 확인됨 (시도 %d/%d)\n", i+1, maxRetries)
					return nil
				}

				// 준비 상태 파일 생성 메시지 확인
				if strings.Contains(logOutput, "✅ 서버 준비 상태 파일 생성됨") {
					fmt.Printf("✅ 집계자 서버 준비 완료! 준비 상태 메시지 확인됨 (시도 %d/%d)\n", i+1, maxRetries)
					return nil
				}

				// ROUND 1 시작 확인 (더 확실한 준비 상태)
				if strings.Contains(logOutput, "[ROUND 1]") {
					fmt.Printf("✅ 집계자 서버 준비 완료! Round 1 시작 확인됨 (시도 %d/%d)\n", i+1, maxRetries)
					return nil
				}

				fmt.Printf("⏳ 아직 준비되지 않음. 다음 시도를 기다립니다...\n")
			}

			// 2. 상태 파일 존재 확인 (백업 방법)
			_, _, fileErr := sshClient.ExecuteCommand("ls ~/federated_learning/server_ready.txt")
			if fileErr == nil {
				fmt.Printf("✅ 집계자 서버 준비 완료! 상태 파일 확인됨 (시도 %d/%d)\n", i+1, maxRetries)
				return nil
			}

			// 3. 프로세스 확인 (추가 백업 방법)
			output, _, procErr := sshClient.ExecuteCommand("pgrep -f 'server_app.py'")
			if procErr == nil && len(strings.TrimSpace(output)) > 0 {
				fmt.Printf("✅ 집계자 서버 준비 완료! server_app.py 프로세스 확인됨 (PID: %s) (시도 %d/%d)\n", strings.TrimSpace(output), i+1, maxRetries)
				return nil
			}

			// 현재 로그 상태 출력 (디버깅용)
			if i%3 == 0 && len(strings.TrimSpace(logOutput)) > 0 { // 3번에 한 번만 출력
				fmt.Printf("📋 현재 로그 상태:\n%s\n", logOutput)
			}
		}

		fmt.Printf("⏳ 집계자 서버 준비 중... (시도 %d/%d)\n", i+1, maxRetries)

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("집계자 서버가 %d초 내에 준비되지 않았습니다", maxRetries*int(retryInterval.Seconds()))
}

// GetMLflowDashboardURL은 연합학습의 MLflow 대시보드 URL을 반환합니다
func (h *FederatedLearningHandler) GetMLflowDashboardURL(c *gin.Context) {
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

	// 집계자 정보 조회
	if fl.AggregatorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "집계자가 설정되지 않았습니다"})
		return
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "집계자 조회에 실패했습니다"})
		return
	}

	if aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "집계자를 찾을 수 없습니다"})
		return
	}

	// MLflow 대시보드 URL 생성 (포트 5000)
	mlflowURL := fmt.Sprintf("http://%s:5000", aggregator.PublicIP)

	response := gin.H{
		"federatedLearningId": fl.ID,
		"aggregatorId":        aggregator.ID,
		"mlflowURL":           mlflowURL,
		"experimentName":      fmt.Sprintf("federated-learning-%s", fl.ID),
		"status":              fl.Status,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetMLflowMetrics는 연합학습의 MLflow 메트릭을 조회합니다
func (h *FederatedLearningHandler) GetMLflowMetrics(c *gin.Context) {
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

	// 집계자 정보 조회
	if fl.AggregatorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "집계자가 설정되지 않았습니다"})
		return
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "집계자 조회에 실패했습니다"})
		return
	}

	if aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "집계자를 찾을 수 없습니다"})
		return
	}

	// MLflow API를 통해 메트릭 조회
	mlflowBaseURL := fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
	experimentName := fmt.Sprintf("federated-learning-%s", fl.ID)

	metrics, err := h.fetchMLflowMetrics(mlflowBaseURL, experimentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("메트릭 조회 실패: %v", err)})
		return
	}

	response := gin.H{
		"federatedLearningId": fl.ID,
		"experimentName":      experimentName,
		"mlflowURL":           mlflowBaseURL,
		"metrics":             metrics,
		"lastUpdated":         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// fetchMLflowMetrics는 MLflow REST API를 통해 메트릭을 조회합니다
func (h *FederatedLearningHandler) fetchMLflowMetrics(mlflowURL, experimentName string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 1. 실험 조회
	experimentURL := fmt.Sprintf("%s/api/2.0/mlflow/experiments/get-by-name?experiment_name=%s", mlflowURL, experimentName)

	resp, err := client.Get(experimentURL)
	if err != nil {
		return nil, fmt.Errorf("실험 조회 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("실험 조회 실패: 상태 코드 %d", resp.StatusCode)
	}

	var experimentResp struct {
		Experiment struct {
			ExperimentID string `json:"experiment_id"`
		} `json:"experiment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&experimentResp); err != nil {
		return nil, fmt.Errorf("실험 응답 파싱 실패: %v", err)
	}

	// 2. 실험의 런 조회
	runsURL := fmt.Sprintf("%s/api/2.0/mlflow/runs/search", mlflowURL)
	searchPayload := map[string]interface{}{
		"experiment_ids": []string{experimentResp.Experiment.ExperimentID},
		"max_results":    1,
	}

	searchData, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, fmt.Errorf("검색 요청 생성 실패: %v", err)
	}

	resp, err = client.Post(runsURL, "application/json", bytes.NewBuffer(searchData))
	if err != nil {
		return nil, fmt.Errorf("런 조회 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("런 조회 실패: 상태 코드 %d", resp.StatusCode)
	}

	var runsResp struct {
		Runs []struct {
			Info struct {
				RunID string `json:"run_id"`
			} `json:"info"`
			Data struct {
				Metrics []struct {
					Key       string  `json:"key"`
					Value     float64 `json:"value"`
					Timestamp int64   `json:"timestamp"`
					Step      int     `json:"step"`
				} `json:"metrics"`
				Params []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"params"`
			} `json:"data"`
		} `json:"runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runsResp); err != nil {
		return nil, fmt.Errorf("런 응답 파싱 실패: %v", err)
	}

	if len(runsResp.Runs) == 0 {
		return map[string]interface{}{
			"metrics": []interface{}{},
			"params":  map[string]interface{}{},
			"status":  "no_runs_found",
		}, nil
	}

	run := runsResp.Runs[0]

	// 메트릭을 스텝별로 그룹화
	metricsMap := make(map[string][]map[string]interface{})
	for _, metric := range run.Data.Metrics {
		if metricsMap[metric.Key] == nil {
			metricsMap[metric.Key] = []map[string]interface{}{}
		}
		metricsMap[metric.Key] = append(metricsMap[metric.Key], map[string]interface{}{
			"step":      metric.Step,
			"value":     metric.Value,
			"timestamp": metric.Timestamp,
		})
	}

	// 파라미터를 맵으로 변환
	paramsMap := make(map[string]interface{})
	for _, param := range run.Data.Params {
		paramsMap[param.Key] = param.Value
	}

	// 3. 런의 전체 메트릭 히스토리 조회 (더 상세한 데이터를 위해)
	runID := run.Info.RunID
	metricsHistoryURL := fmt.Sprintf("%s/api/2.0/mlflow/metrics/get-history?run_id=%s&metric_key=", mlflowURL, runID)

	// 주요 메트릭들에 대한 히스토리 조회
	keyMetrics := []string{"train_loss", "val_loss", "accuracy", "f1_macro"}
	detailedMetrics := make(map[string][]map[string]interface{})

	for _, metricKey := range keyMetrics {
		historyURL := metricsHistoryURL + metricKey
		resp, err := client.Get(historyURL)
		if err != nil {
			continue // 에러가 있어도 다른 메트릭은 계속 조회
		}

		if resp.StatusCode == http.StatusOK {
			var historyResp struct {
				Metrics []struct {
					Key       string  `json:"key"`
					Value     float64 `json:"value"`
					Timestamp int64   `json:"timestamp"`
					Step      int     `json:"step"`
				} `json:"metrics"`
			}

			if json.NewDecoder(resp.Body).Decode(&historyResp) == nil {
				for _, metric := range historyResp.Metrics {
					if detailedMetrics[metricKey] == nil {
						detailedMetrics[metricKey] = []map[string]interface{}{}
					}
					detailedMetrics[metricKey] = append(detailedMetrics[metricKey], map[string]interface{}{
						"step":      metric.Step,
						"value":     metric.Value,
						"timestamp": metric.Timestamp,
					})
				}
			}
		}
		resp.Body.Close()
	}

	result := map[string]interface{}{
		"runId":           runID,
		"metrics":         metricsMap,
		"detailedMetrics": detailedMetrics,
		"params":          paramsMap,
		"status":          "success",
	}

	return result, nil
}

// GetLatestMetrics는 최신 메트릭만 간단히 조회합니다 (폴링용)
func (h *FederatedLearningHandler) GetLatestMetrics(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 연합학습 ID 추출
	id := c.Param("id")

	// DB에서 연합학습 조회
	fl, err := h.repo.GetByID(id)
	if err != nil || fl == nil || fl.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "연합학습 작업을 찾을 수 없습니다"})
		return
	}

	// 집계자 정보 조회
	if fl.AggregatorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "집계자가 설정되지 않았습니다"})
		return
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
	if err != nil || aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "집계자를 찾을 수 없습니다"})
		return
	}

	// MLflow API를 통해 최신 메트릭만 조회
	mlflowBaseURL := fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
	experimentName := fmt.Sprintf("federated-learning-%s", fl.ID)

	latestMetrics, err := h.fetchLatestMLflowMetrics(mlflowBaseURL, experimentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("메트릭 조회 실패: %v", err)})
		return
	}

	response := gin.H{
		"federatedLearningId": fl.ID,
		"status":              fl.Status,
		"metrics":             latestMetrics,
		"timestamp":           time.Now().Unix(),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// fetchLatestMLflowMetrics는 최신 메트릭값들만 빠르게 조회합니다
func (h *FederatedLearningHandler) fetchLatestMLflowMetrics(mlflowURL, experimentName string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // 빠른 응답을 위해 타임아웃 단축
	}

	// 실험 조회
	experimentURL := fmt.Sprintf("%s/api/2.0/mlflow/experiments/get-by-name?experiment_name=%s", mlflowURL, experimentName)

	resp, err := client.Get(experimentURL)
	if err != nil {
		return map[string]interface{}{"status": "mlflow_unavailable"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"status": "experiment_not_found"}, nil
	}

	var experimentResp struct {
		Experiment struct {
			ExperimentID string `json:"experiment_id"`
		} `json:"experiment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&experimentResp); err != nil {
		return map[string]interface{}{"status": "parse_error"}, nil
	}

	// 최신 런의 메트릭 조회
	runsURL := fmt.Sprintf("%s/api/2.0/mlflow/runs/search", mlflowURL)
	searchPayload := map[string]interface{}{
		"experiment_ids": []string{experimentResp.Experiment.ExperimentID},
		"max_results":    1,
		"order_by":       []string{"attribute.start_time DESC"},
	}

	searchData, _ := json.Marshal(searchPayload)

	resp, err = client.Post(runsURL, "application/json", bytes.NewBuffer(searchData))
	if err != nil {
		return map[string]interface{}{"status": "runs_unavailable"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"status": "runs_not_found"}, nil
	}

	var runsResp struct {
		Runs []struct {
			Data struct {
				Metrics []struct {
					Key   string  `json:"key"`
					Value float64 `json:"value"`
					Step  int     `json:"step"`
				} `json:"metrics"`
			} `json:"data"`
		} `json:"runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runsResp); err != nil || len(runsResp.Runs) == 0 {
		return map[string]interface{}{"status": "no_data"}, nil
	}

	// 각 메트릭의 최신값 추출
	latestMetrics := make(map[string]interface{})
	maxSteps := make(map[string]int)

	for _, metric := range runsResp.Runs[0].Data.Metrics {
		if metric.Step >= maxSteps[metric.Key] {
			maxSteps[metric.Key] = metric.Step
			latestMetrics[metric.Key] = map[string]interface{}{
				"value": metric.Value,
				"step":  metric.Step,
			}
		}
	}

	// 진행률 계산 (현재 스텝 기준)
	maxStep := 0
	for _, step := range maxSteps {
		if step > maxStep {
			maxStep = step
		}
	}

	// 예상 총 라운드 (DB에서 가져온 값 사용 가능)
	// 여기서는 간단히 10으로 가정, 실제로는 fl.Rounds 사용
	progress := float64(maxStep) / float64(10) * 100
	if progress > 100 {
		progress = 100
	}

	result := map[string]interface{}{
		"status":          "success",
		"latestMetrics":   latestMetrics,
		"currentRound":    maxStep,
		"totalRounds":     10, // 실제로는 fl.Rounds 사용
		"progressPercent": progress,
	}

	return result, nil
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

// GetFederatedLearningLogs는 연합학습 실행 로그를 조회하는 핸들러입니다
func (h *FederatedLearningHandler) GetFederatedLearningLogs(c *gin.Context) {
	fmt.Printf("=== 로그 조회 요청 시작 ===\n")

	// Authorization 헤더 확인
	authHeader := c.GetHeader("Authorization")
	fmt.Printf("Authorization 헤더: %s\n", authHeader)

	// Context에서 userID 확인
	userIDInterface, exists := c.Get("userID")
	if !exists {
		fmt.Printf("userID가 Context에 없음 - 미들웨어가 실행되지 않았을 수 있음\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다. 미들웨어 오류"})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		fmt.Printf("userID 타입 변환 실패: %T\n", userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 정보가 올바르지 않습니다"})
		return
	}

	fmt.Printf("사용자 ID: %d\n", userID)

	// 경로 매개변수에서 연합학습 ID 추출
	id := c.Param("id")
	fmt.Printf("연합학습 ID: %s\n", id)

	// DB에서 연합학습 조회
	fl, err := h.repo.GetByID(id)
	if err != nil {
		fmt.Printf("연합학습 조회 실패: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "연합학습 작업 조회에 실패했습니다"})
		return
	}
	if fl == nil {
		fmt.Printf("연합학습을 찾을 수 없음: %s\n", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "연합학습 작업을 찾을 수 없습니다"})
		return
	}

	// 작업 소유자 확인
	if fl.UserID != userID {
		fmt.Printf("권한 없음 - 요청자: %d, 소유자: %d\n", userID, fl.UserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 연합학습 작업에 접근할 권한이 없습니다"})
		return
	}

	fmt.Printf("로그 조회 시작 - FL ID: %s, User ID: %d\n", fl.ID, userID)

	// 집계자 로그 조회
	aggregatorLogs, err := h.getAggregatorLogs(fl)
	if err != nil {
		fmt.Printf("집계자 로그 조회 실패: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "집계자 로그 조회에 실패했습니다: " + err.Error()})
		return
	}

	response := map[string]interface{}{
		"federatedLearningId": fl.ID,
		"status":              fl.Status,
		"aggregatorLogs":      aggregatorLogs,
	}

	fmt.Printf("로그 조회 성공\n")
	c.JSON(http.StatusOK, gin.H{"data": response})
}

// StreamFederatedLearningLogs는 연합학습 로그를 실시간으로 스트림하는 핸들러입니다
func (h *FederatedLearningHandler) StreamFederatedLearningLogs(c *gin.Context) {
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

	// SSE 헤더 설정
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 클라이언트가 연결을 끊을 때까지 주기적으로 로그 전송
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// 집계자 로그 조회
			aggregatorLogs, err := h.getAggregatorLogs(fl)
			if err != nil {
				c.SSEvent("error", fmt.Sprintf("집계자 로그 조회 실패: %v", err))
				continue
			}

			// 로그 데이터를 SSE 형태로 전송
			logData := map[string]interface{}{
				"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
				"aggregatorLogs": aggregatorLogs,
			}

			c.SSEvent("logs", logData)
			c.Writer.Flush()
		}
	}
}

// getAggregatorLogs는 집계자의 로그를 조회합니다
func (h *FederatedLearningHandler) getAggregatorLogs(fl *models.FederatedLearning) (map[string]interface{}, error) {
	if fl.AggregatorID == nil {
		return nil, fmt.Errorf("집계자 ID가 설정되지 않았습니다")
	}

	// 집계자 정보 조회
	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
	if err != nil {
		return nil, fmt.Errorf("집계자 조회 실패: %v", err)
	}

	if aggregator.PublicIP == "" {
		return nil, fmt.Errorf("집계자의 Public IP가 설정되지 않았습니다")
	}

	// SSH 키페어 조회
	keypairWithPrivateKey, err := h.sshKeypairService.GetKeypairWithPrivateKey(aggregator.ID)
	if err != nil {
		return nil, fmt.Errorf("집계자 SSH 키페어 조회 실패: %v", err)
	}

	// SSH 클라이언트 생성
	sshClient := utils.NewSSHClient(
		aggregator.PublicIP,
		"22",
		"ubuntu",
		keypairWithPrivateKey.PrivateKey,
	)

	// 작업 디렉토리 경로 (federatedLearningID 사용)
	workDir := fmt.Sprintf("/home/ubuntu/fl-aggregator-%s", fl.ID)
	logFilePath := fmt.Sprintf("%s/flower_server.log", workDir)

	// Flower 서버 로그 조회
	flowerLogCmd := fmt.Sprintf("cat %s 2>/dev/null || echo 'No flower server log found'", logFilePath)
	flowerLogs, _, err := sshClient.ExecuteCommand(flowerLogCmd)
	if err != nil {
		flowerLogs = fmt.Sprintf("로그 조회 실패: %v", err)
	}

	return map[string]interface{}{
		"aggregatorId":   aggregator.ID,
		"aggregatorName": aggregator.Name,
		"publicIP":       aggregator.PublicIP,
		"logFilePath":    logFilePath,
		"flowerLogs":     strings.TrimSpace(flowerLogs),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// SyncMLflowMetricsToDatabase는 MLflow 메트릭을 데이터베이스에 동기화합니다
func (h *FederatedLearningHandler) SyncMLflowMetricsToDatabase(c *gin.Context) {
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

	// 집계자 정보 조회
	if fl.AggregatorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "집계자가 설정되지 않았습니다"})
		return
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*fl.AggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "집계자 조회에 실패했습니다"})
		return
	}

	if aggregator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "집계자를 찾을 수 없습니다"})
		return
	}

	// MLflow에서 메트릭 조회 및 데이터베이스 동기화
	mlflowBaseURL := fmt.Sprintf("http://%s:1", aggregator.PublicIP)
	experimentName := fmt.Sprintf("federated-learning-%s", fl.ID)

	syncResult, err := h.syncMetricsFromMLflowToDB(mlflowBaseURL, experimentName, *fl.AggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("메트릭 동기화 실패: %v", err)})
		return
	}

	response := gin.H{
		"federatedLearningId": fl.ID,
		"aggregatorId":        *fl.AggregatorID,
		"experimentName":      experimentName,
		"syncResult":          syncResult,
		"timestamp":           time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetStoredTrainingHistory는 데이터베이스에 저장된 라운드별 학습 히스토리를 조회합니다
func (h *FederatedLearningHandler) GetStoredTrainingHistory(c *gin.Context) {
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

	// 집계자 정보 조회
	if fl.AggregatorID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "집계자가 설정되지 않았습니다"})
		return
	}

	// 데이터베이스에서 저장된 학습 라운드 조회
	trainingRounds, err := h.aggregatorRepo.GetTrainingRoundsByAggregatorID(*fl.AggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "학습 히스토리 조회에 실패했습니다"})
		return
	}

	// 응답 데이터 구성
	response := gin.H{
		"federatedLearningId": fl.ID,
		"aggregatorId":        *fl.AggregatorID,
		"totalRounds":         len(trainingRounds),
		"trainingHistory":     trainingRounds,
		"lastUpdated":         time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// syncMetricsFromMLflowToDB는 MLflow에서 메트릭을 조회하여 데이터베이스에 저장합니다
func (h *FederatedLearningHandler) syncMetricsFromMLflowToDB(mlflowURL, experimentName, aggregatorID string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 1. 실험 조회
	experimentURL := fmt.Sprintf("%s/api/2.0/mlflow/experiments/get-by-name?experiment_name=%s", mlflowURL, experimentName)

	resp, err := client.Get(experimentURL)
	if err != nil {
		return nil, fmt.Errorf("실험 조회 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("실험 조회 실패: 상태 코드 %d", resp.StatusCode)
	}

	var experimentResp struct {
		Experiment struct {
			ExperimentID string `json:"experiment_id"`
		} `json:"experiment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&experimentResp); err != nil {
		return nil, fmt.Errorf("실험 응답 파싱 실패: %v", err)
	}

	// 2. 실험의 모든 런 조회
	runsURL := fmt.Sprintf("%s/api/2.0/mlflow/runs/search", mlflowURL)
	searchPayload := map[string]interface{}{
		"experiment_ids": []string{experimentResp.Experiment.ExperimentID},
		"max_results":    100, // 충분한 수의 런을 가져오기
		"order_by":       []string{"attribute.start_time DESC"},
	}

	searchData, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, fmt.Errorf("검색 요청 생성 실패: %v", err)
	}

	resp, err = client.Post(runsURL, "application/json", bytes.NewBuffer(searchData))
	if err != nil {
		return nil, fmt.Errorf("런 조회 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("런 조회 실패: 상태 코드 %d", resp.StatusCode)
	}

	var runsResp struct {
		Runs []struct {
			Info struct {
				RunID     string `json:"run_id"`
				StartTime int64  `json:"start_time"`
				EndTime   int64  `json:"end_time"`
			} `json:"info"`
			Data struct {
				Metrics []struct {
					Key       string  `json:"key"`
					Value     float64 `json:"value"`
					Timestamp int64   `json:"timestamp"`
					Step      int     `json:"step"`
				} `json:"metrics"`
			} `json:"data"`
		} `json:"runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runsResp); err != nil {
		return nil, fmt.Errorf("런 응답 파싱 실패: %v", err)
	}

	if len(runsResp.Runs) == 0 {
		return map[string]interface{}{
			"status":      "no_runs_found",
			"savedRounds": 0,
		}, nil
	}

	// 3. 가장 최신 런의 메트릭을 라운드별로 그룹화하여 저장
	latestRun := runsResp.Runs[0]

	// 스텝별 메트릭 그룹화
	stepMetrics := make(map[int]map[string]float64)
	for _, metric := range latestRun.Data.Metrics {
		if stepMetrics[metric.Step] == nil {
			stepMetrics[metric.Step] = make(map[string]float64)
		}
		stepMetrics[metric.Step][metric.Key] = metric.Value
	}

	// 기존 라운드 조회하여 중복 방지
	existingRounds, err := h.aggregatorRepo.GetTrainingRoundsByAggregatorID(aggregatorID)
	if err != nil {
		return nil, fmt.Errorf("기존 라운드 조회 실패: %v", err)
	}

	existingRoundMap := make(map[int]bool)
	for _, round := range existingRounds {
		existingRoundMap[round.Round] = true
	}

	// 4. 각 스텝을 TrainingRound로 변환하여 저장
	savedRounds := 0
	updatedRounds := 0

	for step, metrics := range stepMetrics {
		// 기존 라운드가 있는지 확인
		if existingRoundMap[step] {
			// 업데이트 로직 (필요시)
			updatedRounds++
			continue
		}

		// 새로운 TrainingRound 생성
		trainingRound := &models.TrainingRound{
			ID:           uuid.New().String(),
			AggregatorID: aggregatorID,
			Round:        step,
			ModelMetrics: models.ModelMetric{
				Accuracy:  getFloatPtr(metrics, "accuracy"),
				Loss:      getFloatPtr(metrics, "val_loss"),
				Precision: getFloatPtr(metrics, "precision_macro"),
				Recall:    getFloatPtr(metrics, "recall_macro"),
				F1Score:   getFloatPtr(metrics, "f1_macro"),
			},
			Duration:          120, // 기본값, 실제로는 계산 로직 추가 가능
			ParticipantsCount: 3,   // 기본값, 실제 참가자 수로 업데이트 가능
			StartedAt:         time.Unix(latestRun.Info.StartTime/1000, 0),
		}

		// 완료 시간 설정 (EndTime이 있는 경우)
		if latestRun.Info.EndTime > 0 {
			completedAt := time.Unix(latestRun.Info.EndTime/1000, 0)
			trainingRound.CompletedAt = &completedAt
		}

		// 데이터베이스에 저장
		if err := h.aggregatorRepo.CreateTrainingRound(trainingRound); err != nil {
			return nil, fmt.Errorf("라운드 %d 저장 실패: %v", step, err)
		}

		savedRounds++
	}

	result := map[string]interface{}{
		"status":         "success",
		"totalSteps":     len(stepMetrics),
		"savedRounds":    savedRounds,
		"updatedRounds":  updatedRounds,
		"existingRounds": len(existingRounds),
		"runId":          latestRun.Info.RunID,
		"experimentName": experimentName,
	}

	return result, nil
}

// getFloatPtr은 메트릭 맵에서 float64 포인터를 안전하게 가져옵니다
func getFloatPtr(metrics map[string]float64, key string) *float64 {
	if value, exists := metrics[key]; exists {
		return &value
	}
	return nil
}

// FederatedLearning 생성 응답 구조
type CreateFederatedLearningResponse struct {
	FederatedLearningID string `json:"federatedLearningId"`
	AggregatorID        string `json:"aggregatorId"`
	Status              string `json:"status"`
}
