package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type CloudHandler struct {
	cloudRepo *repository.CloudRepository
}

func NewCloudHandler(cloudRepo *repository.CloudRepository) *CloudHandler {
	return &CloudHandler{cloudRepo: cloudRepo}
}

// GetClouds godoc
// @Summary 클라우드 연결 목록 조회
// @Description 등록된 모든 클라우드 연결 정보를 조회합니다.
// @Tags clouds
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/clouds [get]
func (h *CloudHandler) GetClouds(c *gin.Context) {
	// TODO: 실제 사용자 ID를 세션에서 가져오도록 수정
	userID := int64(1) // 임시로 1번 사용자로 설정

	connections, err := h.cloudRepo.GetCloudConnectionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 연결 목록 조회 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": connections})
}

// AddCloud godoc
// @Summary 클라우드 연결 추가
// @Description 새로운 클라우드 연결을 추가합니다.
// @Tags clouds
// @Accept json
// @Produce json
// @Param cloud body CloudConnection true "클라우드 연결 정보"
// @Success 200 {object} CloudConnection
// @Failure 400 {object} map[string]string
// @Router /api/clouds [post]
func (h *CloudHandler) AddCloud(c *gin.Context) {
	var cloud models.CloudConnection
	if err := c.ShouldBindJSON(&cloud); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다."})
		return
	}

	// TODO: 실제 사용자 ID를 세션에서 가져오도록 수정
	cloud.UserID = 1 // 임시로 1번 사용자로 설정

	// 클라우드 연결 테스트
	if err := testCloudConnection(c.Request.Context(), cloud); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "클라우드 연결 실패: " + err.Error()})
		return
	}

	// 연결 성공 시 저장
	cloud.Status = "connected"
	if err := h.cloudRepo.CreateCloudConnection(&cloud); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 연결 저장 실패"})
		return
	}

	c.JSON(http.StatusOK, cloud)
}

// DeleteCloud godoc
// @Summary 클라우드 연결 삭제
// @Description 지정된 ID의 클라우드 연결을 삭제합니다.
// @Tags clouds
// @Accept json
// @Produce json
// @Param id path string true "클라우드 연결 ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/clouds/{id} [delete]
func (h *CloudHandler) DeleteCloud(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 ID 형식"})
		return
	}

	// TODO: 실제 사용자 ID를 세션에서 가져오도록 수정
	userID := int64(1) // 임시로 1번 사용자로 설정

	// 연결 조회
	conn, err := h.cloudRepo.GetCloudConnectionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 연결 조회 실패"})
		return
	}
	if conn == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "클라우드 연결을 찾을 수 없습니다"})
		return
	}

	// 사용자 확인
	if conn.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	// 연결 삭제
	if err := h.cloudRepo.DeleteCloudConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 연결 삭제 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "클라우드 연결이 성공적으로 삭제되었습니다"})
}

func testCloudConnection(ctx context.Context, cloud models.CloudConnection) error {
	switch cloud.Provider {
	case "AWS":
		if len(cloud.CredentialFile) == 0 {
			return fmt.Errorf("AWS 자격 증명 파일이 필요합니다")
		}

		// AWS 자격 증명 파일 파싱
		creds := string(cloud.CredentialFile)
		lines := strings.Split(creds, "\n")
		var accessKey, secretKey, region string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "aws_access_key_id") {
				accessKey = strings.Split(line, "=")[1]
				accessKey = strings.TrimSpace(accessKey)
			} else if strings.HasPrefix(line, "aws_secret_access_key") {
				secretKey = strings.Split(line, "=")[1]
				secretKey = strings.TrimSpace(secretKey)
			} else if strings.HasPrefix(line, "region") {
				region = strings.Split(line, "=")[1]
				region = strings.TrimSpace(region)
			}
		}

		if accessKey == "" || secretKey == "" || region == "" {
			return fmt.Errorf("AWS 자격 증명 파일에 필수 정보가 누락되었습니다")
		}

		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			)),
		)
		if err != nil {
			return err
		}

		// EC2 클라이언트 생성
		client := ec2.NewFromConfig(cfg)

		// AMI 목록 조회 (Amazon Linux 2)
		amiResult, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name:   stringPtr("name"),
					Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"},
				},
				{
					Name:   stringPtr("state"),
					Values: []string{"available"},
				},
			},
			Owners: []string{"amazon"},
		})
		if err != nil {
			return fmt.Errorf("AMI 목록 조회 실패: %v", err)
		}

		// 인스턴스 타입 목록 조회
		instanceTypes, err := client.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{
			MaxResults: int32Ptr(5),
		})
		if err != nil {
			return fmt.Errorf("인스턴스 타입 목록 조회 실패: %v", err)
		}

		if len(amiResult.Images) == 0 {
			return fmt.Errorf("사용 가능한 AMI를 찾을 수 없습니다")
		}

		if len(instanceTypes.InstanceTypes) == 0 {
			return fmt.Errorf("사용 가능한 인스턴스 타입을 찾을 수 없습니다")
		}

		return nil

	case "GCP":
		if len(cloud.CredentialFile) == 0 {
			return fmt.Errorf("GCP 자격 증명 파일이 필요합니다")
		}

		// 프로젝트 ID 추출
		var creds map[string]interface{}
		if err := json.Unmarshal(cloud.CredentialFile, &creds); err != nil {
			return fmt.Errorf("자격 증명 파일 파싱 실패: %v", err)
		}

		projectID, ok := creds["project_id"].(string)
		if !ok {
			return fmt.Errorf("프로젝트 ID를 찾을 수 없습니다")
		}

		// Compute Engine 서비스 생성
		computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(cloud.CredentialFile))
		if err != nil {
			return fmt.Errorf("gcp Compute Engine 서비스 생성 실패: %v", err)
		}

		// 이미지 목록 조회 (Ubuntu)
		imageList, err := computeService.Images.List(projectID).Filter("name eq ubuntu-*").MaxResults(1).Do()
		if err != nil {
			return fmt.Errorf("이미지 목록 조회 실패: %v", err)
		}

		// 머신 타입 목록 조회
		machineTypes, err := computeService.MachineTypes.List(projectID, cloud.Region).MaxResults(1).Do()
		if err != nil {
			return fmt.Errorf("머신 타입 목록 조회 실패: %v", err)
		}

		if len(imageList.Items) == 0 {
			return fmt.Errorf("사용 가능한 이미지를 찾을 수 없습니다")
		}

		if len(machineTypes.Items) == 0 {
			return fmt.Errorf("사용 가능한 머신 타입을 찾을 수 없습니다")
		}

		return nil

	default:
		return fmt.Errorf("지원하지 않는 클라우드 제공자입니다")
	}
}

// UploadCloudCredential godoc
// @Summary 클라우드 자격 증명 파일 업로드
// @Description AWS 또는 GCP 자격 증명 파일을 업로드합니다.
// @Tags clouds
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "자격 증명 파일"
// @Param provider formData string true "클라우드 제공자 (AWS 또는 GCP)"
// @Param name formData string true "연결 이름"
// @Param region formData string false "AWS 리전 (AWS인 경우 필수)"
// @Param projectId formData string false "GCP 프로젝트 ID (GCP인 경우 필수)"
// @Success 200 {object} CloudConnection
// @Failure 400 {object} map[string]string
// @Router /api/clouds/upload [post]
func (h *CloudHandler) UploadCloudCredential(c *gin.Context) {
	// TODO: 실제 사용자 ID를 세션에서 가져오도록 수정
	userID := int64(1) // 임시로 1번 사용자로 설정

	provider := c.PostForm("provider")
	name := c.PostForm("name")
	region := c.PostForm("region")

	file, err := c.FormFile("credentialFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "인증 파일이 필요합니다"})
		return
	}

	// 파일 읽기
	credentialFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 읽기 실패"})
		return
	}
	defer credentialFile.Close()

	// 파일 크기 확인
	if file.Size > 1024*1024 { // 1MB 제한
		c.JSON(http.StatusBadRequest, gin.H{"error": "파일 크기는 1MB를 초과할 수 없습니다"})
		return
	}

	// 파일 내용 읽기
	buffer := make([]byte, file.Size)
	_, err = credentialFile.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 읽기 실패"})
		return
	}

	// 클라우드 연결 생성
	conn := &models.CloudConnection{
		UserID:         userID,
		Provider:       provider,
		Name:           name,
		Region:         region,
		Status:         "pending",
		CredentialFile: buffer,
	}

	// 연결 테스트
	if err := testCloudConnection(c.Request.Context(), *conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 데이터베이스에 저장
	if err := h.cloudRepo.CreateCloudConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 인증 정보 저장 실패"})
		return
	}

	// 상태 업데이트
	if err := h.cloudRepo.UpdateCloudConnectionStatus(conn.ID, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "클라우드 인증 정보 상태 업데이트 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "클라우드 연결이 성공적으로 추가되었습니다"})
}

// Helper functions for AWS SDK
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}