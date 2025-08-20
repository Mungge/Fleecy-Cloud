package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tfexec "github.com/hashicorp/terraform-exec/tfexec"
)

type TerraformConfig struct {
	WorkingDir    string
	CloudProvider string // aws, gcp
	ProjectName   string
	Region        string
	Zone          string
	InstanceType  string
	Environment   string

	// GCP 전용 (nullable)
	ProjectID *string

	// 추가 공통 설정
	StorageSpecs string
	AggregatorID string
	Algorithm    string

	// 클라우드 자격증명
	AWSAccessKey         string
	AWSSecretKey         string
	GCPServiceAccountKey string
}

type TerraformResult struct {
	InstanceID   string `json:"instance_id"`
	PublicIP     string `json:"public_ip"`
	PrivateIP    string `json:"private_ip"`
	Status       string `json:"status"`
	WorkspaceDir string `json:"workspace_dir"`
}

// CreateTerraformWorkspace creates a unique workspace for the deployment
func CreateTerraformWorkspace(aggregatorID string, config TerraformConfig) (string, error) {
	// 공통 작업 디렉토리 사용 (aggregator ID별로 구분)
	workspaceDir := filepath.Join("/tmp", "terraform-workspaces", aggregatorID)

	// 디렉토리 생성
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %v", err)
	}

	// Terraform 설정 파일들 경로 설정 (프로젝트 루트의 terraform 폴더)
	var sourceDir string
	workingDir, _ := os.Getwd()
	// backend 디렉토리에서 한 단계 위로 올라가서 terraform 폴더 찾기
	projectRoot := filepath.Dir(workingDir)
	switch config.CloudProvider {
	case "aws":
		sourceDir = filepath.Join(projectRoot, "terraform", "aws")
	case "gcp":
		sourceDir = filepath.Join(projectRoot, "terraform", "gcp")
	default:
		sourceDir = filepath.Join(projectRoot, "terraform", "aws")
	}
	
	// 소스 디렉토리 존재 확인
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return "", fmt.Errorf("terraform source directory not found: %s", sourceDir)
	}
	if err := copyTerraformFiles(sourceDir, workspaceDir); err != nil {
		return "", fmt.Errorf("failed to copy terraform files: %v", err)
	}

	// 변수 파일 생성
	if err := createTerraformVars(workspaceDir, aggregatorID, config); err != nil {
		return "", fmt.Errorf("failed to create terraform vars: %v", err)
	}

	return workspaceDir, nil
}

// copyTerraformFiles copies terraform configuration files to workspace
func copyTerraformFiles(sourceDir, destDir string) error {
	files := []string{"main.tf", "variables.tf", "outputs.tf", "providers.tf", "locals.tf"}

	for _, file := range files {
		sourcePath := filepath.Join(sourceDir, file)
		destPath := filepath.Join(destDir, file)

		// 파일이 존재하는지 확인
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue // 파일이 없으면 건너뛰기
		}

		sourceBytes, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %v", file, err)
		}

		if err := os.WriteFile(destPath, sourceBytes, 0644); err != nil {
			return fmt.Errorf("failed to write dest file %s: %v", file, err)
		}
	}

	return nil
}

// createTerraformVars creates terraform.tfvars file with specific configuration
func createTerraformVars(workspaceDir, aggregatorID string, config TerraformConfig) error {
	var varsContent string

	// 클라우드별 변수 생성
	switch strings.ToLower(config.CloudProvider) {
	case "aws":
		if config.AWSAccessKey == "" || config.AWSSecretKey == "" {
			return fmt.Errorf("AWS credentials are required for AWS deployment")
		}
		varsContent = fmt.Sprintf(`aws_region = "%s"
availability_zone = "%s"
project_name = "%s"
environment = "%s"
instance_type = "%s"
aggregator_id = "%s"
aws_access_key = "%s"
aws_secret_key = "%s"
storage_specs = "%s"
algorithm = "%s"
`, config.Region, config.Zone, config.ProjectName, config.Environment, config.InstanceType, 
   aggregatorID, config.AWSAccessKey, config.AWSSecretKey, config.StorageSpecs, config.Algorithm)
	case "gcp":
		projectID := ""
		if config.ProjectID != nil {
			projectID = *config.ProjectID
		}
		if projectID == "" {
			return fmt.Errorf("project ID is required for GCP deployment")
		}
		varsContent = fmt.Sprintf(`project_id = "%s"
project_name = "%s"
region = "%s"
zone = "%s"
environment = "%s"
instance_type = "%s"
aggregator_id = "%s"
storage_specs = "%s"
algorithm = "%s"
`, projectID, config.ProjectName, config.Region, config.Zone, config.Environment, 
   config.InstanceType, aggregatorID, config.StorageSpecs, config.Algorithm)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", config.CloudProvider)
	}

	varsPath := filepath.Join(workspaceDir, "terraform.tfvars")
	if err := os.WriteFile(varsPath, []byte(varsContent), 0644); err != nil {
		return fmt.Errorf("failed to write terraform vars file: %v", err)
	}
	
	fmt.Printf("Created terraform vars file: %s\n", varsPath)
	return nil
}

// DeployWithTerraform executes terraform deployment (uses terraform-exec)
func DeployWithTerraform(workspaceDir string) (*TerraformResult, error) {
	return DeployWithTerraformExec(context.Background(), workspaceDir)
}

// DeployWithTerraformContext executes terraform deployment with context support
func DeployWithTerraformContext(ctx context.Context, workspaceDir string) (*TerraformResult, error) {
	return DeployWithTerraformExec(ctx, workspaceDir)
}

// DeployWithTerraformExec uses HashiCorp terraform-exec to run Terraform
func DeployWithTerraformExec(ctx context.Context, workspaceDir string) (*TerraformResult, error) {
	terraformBinary, err := exec.LookPath("terraform")
	if err != nil {
		return nil, fmt.Errorf("terraform binary not found in PATH: %v", err)
	}

	tf, err := tfexec.NewTerraform(workspaceDir, terraformBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %v", err)
	}

	if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
		return nil, fmt.Errorf("terraform init failed: %v", err)
	}

	if err := tf.Apply(ctx); err != nil {
		return nil, fmt.Errorf("terraform apply failed: %v", err)
	}

	outputs, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform outputs: %v", err)
	}

	getString := func(key string) string {
		if meta, ok := outputs[key]; ok && meta.Value != nil {
			b, _ := json.Marshal(meta.Value)
			var s string
			if err := json.Unmarshal(b, &s); err == nil {
				return s
			}
			return string(b)
		}
		return ""
	}

	result := &TerraformResult{
		Status:       "deployed",
		WorkspaceDir: workspaceDir,
		InstanceID:   getString("instance_id"),
		PublicIP:     getString("public_ip"),
		PrivateIP:    getString("private_ip"),
	}

	return result, nil
}
