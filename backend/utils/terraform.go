package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Terraform 설정 파일들을 현재 디렉토리 기준으로 복사
	var sourceDir string
	switch config.CloudProvider {
	case "aws":
		sourceDir = "../terraform/aws"
	case "gcp":
		sourceDir = "../terraform/gcp"
	default:
		sourceDir = "../terraform/aws"
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
	if config.CloudProvider == "aws" {
		varsContent = fmt.Sprintf(`aws_region = "%s"
project_name = "%s"
environment = "%s"
instance_type = "%s"
aggregator_id = "%s"
aws_access_key = "%s"
aws_secret_key = "%s"
`, config.Region, config.ProjectName, config.Environment, config.InstanceType, aggregatorID, config.AWSAccessKey, config.AWSSecretKey)
	} else if config.CloudProvider == "gcp" {
		projectID := ""
		if config.ProjectID != nil {
			projectID = *config.ProjectID
		}
		varsContent = fmt.Sprintf(`project_id = "%s"
project_name = "%s"
region = "%s"
zone = "%s"
environment = "%s"
instance_type = "%s"
aggregator_id = "%s"
`, projectID, config.ProjectName, config.Region, config.Zone, config.Environment, config.InstanceType, aggregatorID)
	}

	varsPath := filepath.Join(workspaceDir, "terraform.tfvars")
	return os.WriteFile(varsPath, []byte(varsContent), 0644)
}

// DeployWithTerraform executes terraform deployment (uses terraform-exec)
func DeployWithTerraform(workspaceDir string) (*TerraformResult, error) {
	return DeployWithTerraformExec(workspaceDir)
}

// DeployWithTerraformExec uses HashiCorp terraform-exec to run Terraform
func DeployWithTerraformExec(workspaceDir string) (*TerraformResult, error) {
	ctx := context.Background()
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
