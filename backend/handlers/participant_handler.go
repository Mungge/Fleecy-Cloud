package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// ParticipantHandler는 참여자(OpenStack 클라우드) 관련 API 핸들러입니다
type ParticipantHandler struct {
	repo            *repository.ParticipantRepository
	openStackService *services.OpenStackService
}

// NewParticipantHandler는 새 ParticipantHandler 인스턴스를 생성합니다
func NewParticipantHandler(repo *repository.ParticipantRepository) *ParticipantHandler {
	return &ParticipantHandler{
		repo:            repo,
		openStackService: services.NewOpenStackService(),
	}
}

// 새 참여자(OpenStack 클라우드)를 생성하는 핸들러
func (h *ParticipantHandler) CreateParticipant(c *gin.Context) {
	// 사용자 ID 추출
	userID := utils.GetUserIDFromMiddleware(c)

	// 요청 본문 파싱
	var request struct {
		Name                 string `json:"name" binding:"required"`
		Metadata             string `json:"metadata"`
		OpenStackEndpoint    string `json:"openstack_endpoint"`
		OpenStackUsername    string `json:"openstack_username"`
		OpenStackPassword    string `json:"openstack_password"`
		OpenStackProjectName string `json:"openstack_project_name"`
		OpenStackDomainName  string `json:"openstack_domain_name"`
		OpenStackRegion      string `json:"openstack_region"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 참여자 객체 생성
	participant := &models.Participant{
		ID:       uuid.New().String(),
		UserID:   userID,
		Name:     request.Name,
		Status:   "inactive", // 기본적으로 비활성 상태로 생성
		Metadata: request.Metadata,
		
		// OpenStack 관련 필드
		OpenStackEndpoint:    request.OpenStackEndpoint,
		OpenStackUsername:    request.OpenStackUsername,
		OpenStackPassword:    request.OpenStackPassword, // 실제로는 암호화 저장 필요 TODO: 암호화
		OpenStackProjectName: request.OpenStackProjectName,
		OpenStackDomainName:  request.OpenStackDomainName,
		OpenStackRegion:      request.OpenStackRegion,
		
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// DB에 저장
	if err := h.repo.Create(participant); err != nil {
		fmt.Printf("참여자 생성 DB 오류: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "참여자 생성에 실패했습니다", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": participant})
}

// GetParticipants는 사용자의 모든 참여자를 반환하는 핸들러입니다
func (h *ParticipantHandler) GetParticipants(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	// 사용자의 모든 참여자 조회
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

	// 요청 본문 파싱
	var request struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Metadata string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	// 필드 업데이트
	if request.Name != "" {
		participant.Name = request.Name
	}
	if request.Status != "" {
		participant.Status = request.Status
	}
	if request.Metadata != "" {
		participant.Metadata = request.Metadata
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
