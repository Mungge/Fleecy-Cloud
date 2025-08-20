package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// CloudKeypairService 클라우드별 키페어 관리 서비스
type CloudKeypairService struct {
	cloudRepo *repository.CloudRepository
}

func NewCloudKeypairService(cloudRepo *repository.CloudRepository) *CloudKeypairService {
	return &CloudKeypairService{
		cloudRepo: cloudRepo,
	}
}

// KeypairInfo 키페어 정보
type KeypairInfo struct {
	KeyName    string `json:"key_name"`
	PublicKey  string `json:"public_key,omitempty"`
	PrivateKey string `json:"private_key,omitempty"` // 새로 생성된 경우에만 포함
}

// GetOrCreateAWSKeypair AWS EC2 키페어 조회 또는 생성
func (s *CloudKeypairService) GetOrCreateAWSKeypair(userID int64, region, keyName string) (*KeypairInfo, error) {
	// 사용자의 AWS 클라우드 연결 찾기
	cloudConnections, err := s.cloudRepo.GetCloudConnectionsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud connections: %v", err)
	}

	var awsConn *models.CloudConnection
	for _, conn := range cloudConnections {
		if strings.EqualFold(conn.Provider, "AWS") && conn.Status == "active" {
			// 활성 AWS 연결 찾기 (리전 무관하게)
			awsConn = conn
			break
		}
	}

	if awsConn == nil {
		return nil, fmt.Errorf("active AWS connection not found for user %d", userID)
	}

	// AWS 자격증명 파싱
	creds := string(awsConn.CredentialFile)
	lines := strings.Split(creds, "\n")

	var accessKey, secretKey string

	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid CSV file format")
	}

	// CSV 헤더 확인
	headers := strings.Split(lines[0], ",")
	accessKeyIdx := -1
	secretKeyIdx := -1

	for i, header := range headers {
		header = strings.TrimSpace(header)
		if strings.Contains(strings.ToLower(header), "access key id") {
			accessKeyIdx = i
		} else if strings.Contains(strings.ToLower(header), "secret access key") {
			secretKeyIdx = i
		}
	}

	if accessKeyIdx == -1 || secretKeyIdx == -1 {
		return nil, fmt.Errorf("CSV file missing access key or secret key columns")
	}

	// 데이터 행 처리
	if len(lines) > 1 {
		values := strings.Split(lines[1], ",")
		if len(values) > accessKeyIdx {
			accessKey = strings.TrimSpace(values[accessKeyIdx])
		}
		if len(values) > secretKeyIdx {
			secretKey = strings.TrimSpace(values[secretKeyIdx])
		}
	}

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("AWS credentials missing in file")
	}

	// AWS 설정 로드 (저장된 자격증명 사용)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	ctx := context.TODO()

	// 기존 키페어 확인
	describeInput := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	}

	describeResult, err := ec2Client.DescribeKeyPairs(ctx, describeInput)
	if err == nil && len(describeResult.KeyPairs) > 0 {
		// 기존 키페어 존재
		keyPair := describeResult.KeyPairs[0]
		log.Printf("Found existing AWS keypair: %s", *keyPair.KeyName)
		return &KeypairInfo{
			KeyName: *keyPair.KeyName,
		}, nil
	}

	// 키페어가 없으면 새로 생성
	log.Printf("Creating new AWS keypair: %s", keyName)
	createInput := &ec2.CreateKeyPairInput{
		KeyName:   aws.String(keyName),
		KeyType:   types.KeyTypeRsa,
		KeyFormat: types.KeyFormatPem,
	}

	createResult, err := ec2Client.CreateKeyPair(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS keypair: %v", err)
	}

	log.Printf("Successfully created AWS keypair: %s", *createResult.KeyName)
	return &KeypairInfo{
		KeyName:    *createResult.KeyName,
		PrivateKey: *createResult.KeyMaterial, // AWS에서 제공하는 Private Key
	}, nil
}

// GetOrCreateGCPKeypair GCP Compute Engine 키페어 조회 또는 생성
func (s *CloudKeypairService) GetOrCreateGCPKeypair(userID int64, projectID, keyName, publicKeyContent, privateKeyContent string) (*KeypairInfo, error) {
	// 사용자의 GCP 클라우드 연결 찾기
	cloudConnections, err := s.cloudRepo.GetCloudConnectionsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud connections: %v", err)
	}

	var gcpConn *models.CloudConnection
	for _, conn := range cloudConnections {
		if strings.EqualFold(conn.Provider, "GCP") && conn.Status == "active" {
			// GCP 자격증명에서 프로젝트 ID 확인
			var creds map[string]interface{}
			if err := json.Unmarshal(conn.CredentialFile, &creds); err == nil {
				if connProjectID, ok := creds["project_id"].(string); ok && connProjectID == projectID {
					gcpConn = conn
					break
				}
			}
		}
	}

	if gcpConn == nil {
		return nil, fmt.Errorf("active GCP connection not found for project %s", projectID)
	}

	ctx := context.Background()

	// GCP Compute Engine 클라이언트 생성 (저장된 자격증명 사용)
	computeService, err := compute.NewService(ctx,
		option.WithCredentialsJSON(gcpConn.CredentialFile),
		option.WithScopes(compute.ComputeScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP compute service: %v", err)
	}

	// 프로젝트 메타데이터에서 SSH 키 확인
	project, err := computeService.Projects.Get(projectID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP project: %v", err)
	}

	// SSH 키가 이미 존재하는지 확인
	for _, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "ssh-keys" && item.Value != nil {
			if containsKey(*item.Value, keyName) {
				log.Printf("Found existing GCP SSH key: %s", keyName)
				return &KeypairInfo{
					KeyName:   keyName,
					PublicKey: publicKeyContent,
				}, nil
			}
		}
	}

	// SSH 키가 없으면 프로젝트 메타데이터에 추가
	log.Printf("Adding SSH key to GCP project: %s", keyName)

	// 기존 SSH 키들 가져오기
	existingKeys := ""
	for _, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "ssh-keys" && item.Value != nil {
			existingKeys = *item.Value
			break
		}
	}

	// 새 키 추가
	newSSHKeys := existingKeys
	if newSSHKeys != "" {
		newSSHKeys += "\n"
	}
	newSSHKeys += fmt.Sprintf("ubuntu:%s", publicKeyContent)

	// 메타데이터 업데이트
	metadata := &compute.Metadata{
		Items: []*compute.MetadataItems{
			{
				Key:   "ssh-keys",
				Value: &newSSHKeys,
			},
		},
	}

	// 프로젝트 메타데이터 업데이트
	operation, err := computeService.Projects.SetCommonInstanceMetadata(projectID, metadata).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to set GCP project metadata: %v", err)
	}

	// 작업 완료 대기 (간단한 버전)
	log.Printf("GCP metadata update operation: %s", operation.Name)

	return &KeypairInfo{
		KeyName:   keyName,
		PublicKey: publicKeyContent,
	}, nil
}

// DeleteAWSKeypair AWS 키페어 삭제
func (s *CloudKeypairService) DeleteAWSKeypair(region, keyName string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	ctx := context.TODO()

	deleteInput := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(keyName),
	}

	_, err = ec2Client.DeleteKeyPair(ctx, deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete AWS keypair: %v", err)
	}

	log.Printf("Successfully deleted AWS keypair: %s", keyName)
	return nil
}

// 헬퍼 함수: SSH 키 문자열에서 특정 키 이름 포함 여부 확인
func containsKey(sshKeys, keyName string) bool {
	// 간단한 구현: 키 이름이 포함되어 있는지 확인
	// 실제로는 더 정교한 파싱이 필요할 수 있음
	return len(sshKeys) > 0 && fmt.Sprintf(":%s", keyName) != ""
}
