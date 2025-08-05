package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type TerraformConfig struct {
	WorkingDir   string
	Region       string
	InstanceType string
	ProjectName  string
	Environment  string
}

type TerraformResult struct {
	InstanceID    string `json:"instance_id"`
	PublicIP      string `json:"public_ip"`
	PrivateIP     string `json:"private_ip"`
	Status        string `json:"status"`
	WorkspaceDir  string `json:"workspace_dir"`
}

// CreateTerraformWorkspace creates a unique workspace for the deployment
func CreateTerraformWorkspace(aggregatorID string, config TerraformConfig) (string, error) {
	// 고유한 작업 디렉토리 생성
	timestamp := time.Now().Unix()
	workspaceDir := filepath.Join("/tmp", fmt.Sprintf("terraform-aggregator-%s-%d", aggregatorID, timestamp))
	
	// 디렉토리 생성
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %v", err)
	}
	
	// Terraform 설정 파일들을 작업 디렉토리로 복사
	sourceDir := "/home/jinhyeok/dev/Fleecy-Cloud/terraform/aws"
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
	files := []string{"main.tf", "variables.tf", "outputs.tf", "providers.tf"}
	
	for _, file := range files {
		sourcePath := filepath.Join(sourceDir, file)
		destPath := filepath.Join(destDir, file)
		
		// 파일이 존재하는지 확인
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue // 파일이 없으면 건너뛰기
		}
		
		sourceBytes, err := ioutil.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %v", file, err)
		}
		
		if err := ioutil.WriteFile(destPath, sourceBytes, 0644); err != nil {
			return fmt.Errorf("failed to write dest file %s: %v", file, err)
		}
	}
	
	return nil
}

// createTerraformVars creates terraform.tfvars file with specific configuration
func createTerraformVars(workspaceDir, aggregatorID string, config TerraformConfig) error {
	varsContent := fmt.Sprintf(`aws_region = "%s"
project_name = "%s"
environment = "%s"
instance_type = "%s"
aggregator_id = "%s"
`, config.Region, config.ProjectName, config.Environment, config.InstanceType, aggregatorID)
	
	varsPath := filepath.Join(workspaceDir, "terraform.tfvars")
	return ioutil.WriteFile(varsPath, []byte(varsContent), 0644)
}

// DeployWithTerraform executes terraform deployment
func DeployWithTerraform(workspaceDir string) (*TerraformResult, error) {
	// Terraform init
	if err := runTerraformCommand(workspaceDir, "init"); err != nil {
		return nil, fmt.Errorf("terraform init failed: %v", err)
	}
	
	// Terraform plan
	if err := runTerraformCommand(workspaceDir, "plan"); err != nil {
		return nil, fmt.Errorf("terraform plan failed: %v", err)
	}
	
	// Terraform apply
	if err := runTerraformCommand(workspaceDir, "apply", "-auto-approve"); err != nil {
		return nil, fmt.Errorf("terraform apply failed: %v", err)
	}
	
	// Get outputs
	return getTerraformOutputs(workspaceDir)
}

// runTerraformCommand executes a terraform command in the specified directory
func runTerraformCommand(workspaceDir string, args ...string) error {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = workspaceDir
	
	// 환경변수 설정 (AWS credentials 등)
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}

// getTerraformOutputs retrieves outputs from terraform state
func getTerraformOutputs(workspaceDir string) (*TerraformResult, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = workspaceDir
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform outputs: %v", err)
	}
	
	// Parse JSON output from Terraform
	var terraformOutput map[string]struct {
		Value string `json:"value"`
	}
	
	if err := json.Unmarshal(output, &terraformOutput); err != nil {
		return nil, fmt.Errorf("failed to parse terraform output JSON: %v", err)
	}
	
	result := &TerraformResult{
		Status:       "deployed",
		WorkspaceDir: workspaceDir,
		InstanceID:   terraformOutput["instance_id"].Value,
		PublicIP:     terraformOutput["public_ip"].Value,
		PrivateIP:    terraformOutput["private_ip"].Value,
	}
	
	return result, nil
}

// Removed generateRandomString function as it is no longer needed.

// CleanupTerraformWorkspace destroys resources and cleans up workspace
func CleanupTerraformWorkspace(workspaceDir string) error {
	// Terraform destroy
	if err := runTerraformCommand(workspaceDir, "destroy", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform destroy failed: %v", err)
	}
	
	// 작업 디렉토리 삭제
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %v", err)
	}
	
	return nil
}
