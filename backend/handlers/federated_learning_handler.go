package handlers

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
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

	// 포트 9092 고정

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
	if federatedLearning.AggregatorID == nil {
		fmt.Printf("집계자 ID가 설정되지 않았습니다\n")
		return
	}

	aggregator, err := h.aggregatorRepo.GetAggregatorByID(*federatedLearning.AggregatorID)
	if err != nil {
		fmt.Printf("집계자 조회 실패: %v\n", err)
		return
	}

	if aggregator == nil {
		fmt.Printf("집계자를 찾을 수 없습니다\n")
		return
	}

	if err := h.sendExecuteRequestToAggregator(aggregator, federatedLearning); err != nil {
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

	// VM 선택 서비스를 사용하여 최적의 VM 선택
	openStackService := services.NewOpenStackService("http://localhost:9090") // Prometheus URL (기본값)
	vmSelectionService := services.NewVMSelectionService(openStackService)
	
	// VM 선택 기준 설정
	criteria := services.VMSelectionCriteria{
		MinVCPUs:       2,
		MinRAM:         4096, // 4GB
		MinDisk:        20,   // 20GB
		MaxCPUUsage:    80.0,
		MaxMemoryUsage: 80.0,
		RequiredStatus: "ACTIVE",
	}
	
	// 최적의 VM 선택
	vmResult, err := vmSelectionService.SelectOptimalVM(participant, criteria)
	if err != nil {
		return fmt.Errorf("VM 선택 실패: %v", err)
	}
	
	if vmResult.SelectedVM == nil {
		return fmt.Errorf("조건을 만족하는 VM을 찾을 수 없습니다: %s", vmResult.SelectionReason)
	}

	// 집계자 주소 가져오기
	aggregatorAddress, err := h.getAggregatorAddress(federatedLearning)
	if err != nil {
		return fmt.Errorf("집계자 주소 조회 실패: %v", err)
	}

	// 참여자용 pyproject.toml 생성 (클라이언트 전용 설정)
	participantPyprojectContent := `[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "flower-client"
version = "1.0.0"
description = "Flower client for federated learning"
license = "Apache-2.0"
dependencies = [
    "flwr[simulation]>=1.20.0",
    "flwr-datasets[vision]>=0.5.0",
    "torch==2.7.1",
    "torchvision==0.22.1",
]

[tool.flwr.app]
publisher = "jinhyeok"

[tool.flwr.app.components]
serverapp = "server_app:app"
clientapp = "client_app:app"

[tool.flwr.federations]
default = "remote-federation"

[tool.flwr.federations.remote-federation]
address = "` + aggregatorAddress + `"
insecure = true
`

	// Flask 서버가 기대하는 페이로드 구성 (파일 내용 포함)
	payload := map[string]interface{}{
		"vm_id": vmResult.SelectedVM.InstanceID,
		"env_config": map[string]interface{}{
			"remote-address": aggregatorAddress,
			"EPOCHS":         10,
			"LR":             0.01,
			"BATCH_SIZE":     32,
			"NUM_ROUNDS":     federatedLearning.Rounds,
			"CLIENT_ID":      participant.ID,
		},
		"files": map[string]interface{}{
			"pyproject.toml": participantPyprojectContent,
			"server_app.py":  serverAppTemplate,
			"client_app.py":  clientAppTemplate,
			"task.py":        taskTemplate,
		},
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
			fmt.Printf("참여자 %s의 엔드포인트가 설정되지 않았습니다\n", participantData.Name)
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

// sendExecuteRequestToAggregator는 집계자에게 SSH를 통해 연합학습 실행 요청을 보냅니다
func (h *FederatedLearningHandler) sendExecuteRequestToAggregator(aggregator *models.Aggregator, federatedLearning *models.FederatedLearning) error {
	// 집계자 Public IP 확인
	if aggregator.PublicIP == "" {
		return fmt.Errorf("집계자 %s의 Public IP가 설정되지 않았습니다", aggregator.Name)
	}

	// SSH 키페어 조회
	keypairWithPrivateKey, err := h.sshKeypairService.GetKeypairWithPrivateKey(aggregator.ID)
	if err != nil {
		return fmt.Errorf("집계자 %s의 SSH 키페어 조회 실패: %v", aggregator.Name, err)
	}

	// SSH 클라이언트 생성
	sshClient := utils.NewSSHClient(
		aggregator.PublicIP,
		"22",
		"ubuntu",
		keypairWithPrivateKey.PrivateKey,
	)

	// SSH 연결 테스트
	if err := sshClient.CheckConnection(); err != nil {
		return fmt.Errorf("집계자 %s SSH 연결 실패: %v", aggregator.Name, err)
	}

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
		return fmt.Errorf("flower 설정 파일 업로드 실패: %v", err)
	}

	// flower_demo 패키지 디렉토리 생성
	_, _, err = sshClient.ExecuteCommand(fmt.Sprintf("mkdir -p %s/flower_demo", workDir))
	if err != nil {
		return fmt.Errorf("flower_demo 디렉토리 생성 실패: %v", err)
	}

	// __init__.py 파일 생성 (Python 패키지로 만들기 위해)
	err = sshClient.UploadFileContent("", fmt.Sprintf("%s/flower_demo/__init__.py", workDir))
	if err != nil {
		return fmt.Errorf("__init__.py 파일 업로드 실패: %v", err)
	}

	// server_app.py 파일을 flower_demo 폴더에 업로드
	err = sshClient.UploadFileContent(serverAppTemplate, fmt.Sprintf("%s/flower_demo/server_app.py", workDir))
	if err != nil {
		return fmt.Errorf("서버 앱 파일 업로드 실패: %v", err)
	}

	// task.py 파일 업로드
	err = sshClient.UploadFileContent(taskTemplate, fmt.Sprintf("%s/flower_demo/task.py", workDir))
	if err != nil {
		return fmt.Errorf("task.py 파일 업로드 실패: %v", err)
	}

	// client_app.py 파일을 flower_demo 폴더에 업로드
	err = sshClient.UploadFileContent(clientAppTemplate, fmt.Sprintf("%s/flower_demo/client_app.py", workDir))
	if err != nil {
		return fmt.Errorf("클라이언트 앱 파일 업로드 실패: %v", err)
	}

	// Flower 서버 실행 스크립트 생성
	runScript := `#!/bin/bash
echo "=== Flower 서버 설정 시작 ==="

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

# 필수 Python 패키지 설치
echo "필수 Python 패키지를 설치합니다..."
pip install flwr torch torchvision

# 설치된 패키지 확인
echo "설치된 패키지 확인:"
pip list | grep -E "(flwr|torch)"

# flwr 명령어 경로 확인
echo "flwr 명령어 경로: $(which flwr)"

echo "Python 패키지 설치가 완료되었습니다."

# Flower 서버 실행 (가상환경에서)
echo "Flower 서버를 시작합니다..."
echo "라운드 수: ` + fmt.Sprintf("%d", federatedLearning.Rounds) + `"
echo "포트: 9092"

# 가상환경에서 flwr 실행 (포트 9092 고정)
source venv/bin/activate && flwr run --run-config num-server-rounds=` + fmt.Sprintf("%d", federatedLearning.Rounds) + ` --superlink 0.0.0.0:9092
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

	fmt.Printf("집계자 %s에서 Flower 서버 실행 성공: %s\n", aggregator.Name, stdout)
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
		"status":             fl.Status,
		"aggregatorLogs":     aggregatorLogs,
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

// FederatedLearning 생성 응답 구조
type CreateFederatedLearningResponse struct {
	FederatedLearningID string `json:"federatedLearningId"`
	AggregatorID        string `json:"aggregatorId"`
	Status              string `json:"status"`
}
