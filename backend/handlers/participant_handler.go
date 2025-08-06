package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

type OpenStackConfig struct {
	Clouds struct {
		OpenStack struct {
			Auth struct {
				AuthURL                     string `yaml:"auth_url"`
				ApplicationCredentialID     string `yaml:"application_credential_id"`
				ApplicationCredentialSecret string `yaml:"application_credential_secret"`
			} `yaml:"auth"`
			RegionName string `yaml:"region_name"`
		} `yaml:"openstack"`
	} `yaml:"clouds"`
}

// ParticipantHandler는 참여자(OpenStack 클라우드) 관련 API 핸들러입니다
type ParticipantHandler struct {
	repo            *repository.ParticipantRepository
	openStackService *services.OpenStackService
}

func NewParticipantHandler(repo *repository.ParticipantRepository) *ParticipantHandler {
	return &ParticipantHandler{
		repo:            repo,
		openStackService: services.NewOpenStackService(),
	}
}

// parseOpenStackConfig는 YAML 파일 내용을 파싱하여 OpenStack 설정을 반환합니다
func (h *ParticipantHandler) parseOpenStackConfig(yamlContent []byte) (*OpenStackConfig, error) {
	var config OpenStackConfig
	if err := yaml.Unmarshal(yamlContent, &config); err != nil {
		return nil, fmt.Errorf("YAML 파싱 오류: %v", err)
	}

	// 필수 필드 검증
	if config.Clouds.OpenStack.Auth.AuthURL == "" {
		return nil, fmt.Errorf("OpenStack auth_url이 누락되었습니다")
	}
	
	// Application Credential 정보가 있어야 함
	if config.Clouds.OpenStack.Auth.ApplicationCredentialID == "" || 
	   config.Clouds.OpenStack.Auth.ApplicationCredentialSecret == "" {
		return nil, fmt.Errorf("application credential ID와 Secret이 필요합니다")
	}

	return &config, nil
}

// CreateParticipant는 새 참여자를 생성하는 핸들러입니다
func (h *ParticipantHandler) CreateParticipant(c *gin.Context) {
	// 사용자 ID 추출
	userID := utils.GetUserIDFromMiddleware(c)

	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "참여자 이름은 필수입니다"})
		return
	}
	region := c.PostForm("region")
	if region == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "리전은 필수입니다"})
		return
	}
	metadata := c.PostForm("metadata")

	// 업로드된 파일 처리
	file, header, err := c.Request.FormFile("configFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "설정 파일이 필요합니다"})
		return
	}
	defer file.Close()

	// 파일 확장자 검증
	filename := header.Filename
	if !(len(filename) > 5 && (filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "YAML 파일만 업로드 가능합니다 (.yaml 또는 .yml)"})
		return
	}

	// 파일 내용 읽기
	yamlContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 읽기에 실패했습니다"})
		return
	}

	// YAML 파싱
	config, err := h.parseOpenStackConfig(yamlContent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// OpenStack endpoint에서 /identity 부분 제거
	authURL := strings.TrimSuffix(config.Clouds.OpenStack.Auth.AuthURL, "/identity")

	// 참여자 객체 생성
	participant := &models.Participant{
		ID:       uuid.New().String(),
		UserID:   userID,
		Name:     name,
		Status:   "inactive", // 기본적으로 비활성 상태로 생성
		Metadata: metadata,
		Region:   region,
		
		// OpenStack 관련 필드
		OpenStackEndpoint:    authURL,
		OpenStackRegion:      config.Clouds.OpenStack.RegionName,
		OpenStackApplicationCredentialID:     config.Clouds.OpenStack.Auth.ApplicationCredentialID,
		OpenStackApplicationCredentialSecret: config.Clouds.OpenStack.Auth.ApplicationCredentialSecret,
		
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// OpenStack 클라우드 연결 테스트
	if err := h.testOpenStackConnection(participant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OpenStack 연결 테스트 실패: " + err.Error()})
		return
	}

	// 연결 테스트 성공 시 활성화
	participant.Status = "active"

	// 테스트 성공 시 DB에 저장
	if err := h.repo.Create(participant); err != nil {
		fmt.Printf("참여자 생성 DB 오류: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 생성에 실패했습니다", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": participant})
}

// GetParticipants는 등록된 모든 참여자를 반환하는 핸들러입니다
func (h *ParticipantHandler) GetParticipants(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 등록된 모든 참여자 조회
	participants, err := h.repo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 목록 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participants})
}

// GetAvailableParticipants는 사용 가능한 참여자 목록을 반환하는 핸들러입니다
func (h *ParticipantHandler) GetAvailableParticipants(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 사용 가능한 참여자 조회
	participants, err := h.repo.GetAvailable(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용 가능한 참여자 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participants})
}

// GetParticipant는 특정 ID의 참여자를 반환하는 핸들러입니다
func (h *ParticipantHandler) GetParticipant(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 참여자 ID 추출
	id := c.Param("id")

	// DB에서 참여자 조회
	participant, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 조회에 실패했습니다"})
		return
	}
	if participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}

	// 참여자 소유자 확인
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 참여자에 접근할 권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participant})
}

func (h *ParticipantHandler) UpdateParticipant(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 참여자 ID 추출
	id := c.Param("id")

	// DB에서 참여자 조회
	participant, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 조회에 실패했습니다"})
		return
	}
	if participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}

	// 참여자 소유자 확인
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 참여자를 수정할 권한이 없습니다"})
		return
	}

	// 폼 데이터에서 name, region, metadata 추출
	if name := c.PostForm("name"); name != "" {
		participant.Name = name
	}
	if region := c.PostForm("region"); region != "" {
		participant.Region = region
	}
	if metadata := c.PostForm("metadata"); metadata != "" {
		participant.Metadata = metadata
	}

	// 업로드된 파일 처리
	file, header, err := c.Request.FormFile("configFile")
	if err == nil {
		// 파일이 업로드된 경우
		defer file.Close()

		// 파일 확장자 검증
		filename := header.Filename
		if !(len(filename) > 5 && (filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "YAML 파일만 업로드 가능합니다 (.yaml 또는 .yml)"})
			return
		}

		// 파일 내용 읽기
		yamlContent, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 읽기에 실패했습니다"})
			return
		}

		// YAML 파싱
		config, err := h.parseOpenStackConfig(yamlContent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// OpenStack 설정 업데이트
		participant.OpenStackEndpoint = config.Clouds.OpenStack.Auth.AuthURL
		participant.OpenStackRegion = config.Clouds.OpenStack.RegionName
		
		// Application Credential 정보 업데이트
		participant.OpenStackApplicationCredentialID = config.Clouds.OpenStack.Auth.ApplicationCredentialID
		participant.OpenStackApplicationCredentialSecret = config.Clouds.OpenStack.Auth.ApplicationCredentialSecret
		
		// OpenStack 설정이 변경된 경우 연결 테스트 실행
		if err := h.testOpenStackConnection(participant); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OpenStack 연결 테스트 실패: " + err.Error()})
			return
		}
		
		// 연결 테스트 성공 시 상태를 active로 변경
		participant.SetActive()
	}

	participant.UpdatedAt = time.Now()

	// DB 업데이트
	if err := h.repo.Update(participant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 업데이트에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participant})
}

func (h *ParticipantHandler) DeleteParticipant(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 경로 매개변수에서 참여자 ID 추출
	id := c.Param("id")

	// DB에서 참여자 조회
	participant, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 조회에 실패했습니다"})
		return
	}
	if participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}

	// 참여자 소유자 확인
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "해당 참여자를 삭제할 권한이 없습니다"})
		return
	}

	// DB에서 삭제
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 삭제에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "참여자가 삭제되었습니다"})
}


// testOpenStackConnection은 OpenStack 클라우드 연결을 테스트합니다
func (h *ParticipantHandler) testOpenStackConnection(participant *models.Participant) error {
	// OpenStack 인증 토큰 획득 테스트
	_, err := h.openStackService.GetAuthToken(participant)
	if err != nil {
		return fmt.Errorf("OpenStack 인증 실패: %v", err)
	}
	
	return nil
}

// HealthCheckParticipant는 참여자의 OpenStack 클라우드 연결 상태를 확인합니다
func (h *ParticipantHandler) HealthCheckParticipant(c *gin.Context) {
	// 경로 매개변수에서 참여자 ID 추출
	id := c.Param("id")

	// DB에서 참여자 조회
	participant, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 조회에 실패했습니다"})
		return
	}
	if participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}

	// OpenStack 연결 테스트
	startTime := time.Now()
	err = h.testOpenStackConnection(participant)
	responseTime := time.Since(startTime).Milliseconds()
	
	healthy := err == nil
	status := "ACTIVE"
	message := fmt.Sprintf("%s 클러스터가 정상적으로 연결되었습니다", participant.Name)
	
	// 참여자 상태 업데이트
	var participantStatusChanged = false
	
	if healthy {
		// 헬스체크 성공 시 active 상태로 변경
		if participant.Status != "active" {
			participant.Status = "active"
			participantStatusChanged = true
		}
	} else {
		// 헬스체크 실패 시 inactive 상태로 변경
		status = "INACTIVE"
		message = fmt.Sprintf("%s OpenStack 연결 실패: %v", participant.Name, err)
		
		if participant.Status != "inactive" {
			participant.Status = "inactive"
			participantStatusChanged = true
		}
	}

	// 상태가 변경된 경우 DB 업데이트
	if participantStatusChanged {
		participant.UpdatedAt = time.Now()
		if updateErr := h.repo.Update(participant); updateErr != nil {
			fmt.Printf("참여자 상태 업데이트 오류: %v\n", updateErr)
			// 상태 업데이트 실패는 로그만 남기고 헬스체크 결과는 반환
		}
	}

	result := map[string]interface{}{
		"healthy":           healthy,
		"status":           status,
		"message":          message,
		"checked_at":       time.Now().Format(time.RFC3339),
		"response_time_ms": responseTime,
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
