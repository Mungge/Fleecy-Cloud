package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// 최적화 요청 구조체
type OptimizationRequest struct {
	FederatedLearning struct {
		Name         string        `json:"name"`
		Description  string        `json:"description"`
		ModelType    string        `json:"modelType"`
		Algorithm    string        `json:"algorithm"`
		Rounds       int           `json:"rounds"`
		Participants []Participant `json:"participants"`
		ModelFileName *string      `json:"modelFileName,omitempty"`
	} `json:"federatedLearning"`
	AggregatorConfig struct {
		MaxBudget  int `json:"maxBudget"`
		MaxLatency int `json:"maxLatency"`
	} `json:"aggregatorConfig"`
}

// 참여자 구조체
type Participant struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	Region            string `json:"openstack_region"`
	OpenstackEndpoint string `json:"openstack_endpoint"`
}

// 최적화 응답 구조체
type OptimizationResponse struct {
	Status           string            `json:"status"`
	Summary          OptimizationSummary `json:"summary"`
	OptimizedOptions []AggregatorOption `json:"optimizedOptions"`
	Message          string            `json:"message"`
	ExecutionTime    float64           `json:"executionTime,omitempty"` // Go에서 추가
}

// 최적화 요약 정보
type OptimizationSummary struct {
	TotalParticipants      int      `json:"totalParticipants"`
	ParticipantRegions     []string `json:"participantRegions"`
	TotalCandidateOptions  int      `json:"totalCandidateOptions"`
	FeasibleOptions        int      `json:"feasibleOptions"`
	Constraints            interface{} `json:"constraints"`
	ModelInfo              interface{} `json:"modelInfo"`
}

// 집계자 옵션 (새로운 구조)
type AggregatorOption struct {
	Rank                   int     `json:"rank"`
	Region                 string  `json:"region"`
	InstanceType           string  `json:"instanceType"`
	CloudProvider          string  `json:"cloudProvider"`
	EstimatedMonthlyCost   float64 `json:"estimatedMonthlyCost"`
	EstimatedHourlyPrice   float64 `json:"estimatedHourlyPrice"`
	AvgLatency             float64 `json:"avgLatency"`
	MaxLatency             float64 `json:"maxLatency"`
	VCPU                   int     `json:"vcpu"`
	Memory                 int     `json:"memory"`
	RecommendationScore    float64 `json:"recommendationScore"`
}

// 최적화 결과 구조체
type OptimizationResult struct {
	Rank            int     `json:"rank"`
	Region          string  `json:"region"`
	InstanceType    string  `json:"instanceType"`
	EstimatedCost   float64 `json:"estimatedCost"`
	EstimatedLatency float64 `json:"estimatedLatency"`
	CloudProvider   string  `json:"cloudProvider"`
}

// OptimizationService 구조체
type OptimizationService struct {
	pythonScriptPath string
	inputDir         string
	outputDir        string
}

// 새로운 OptimizationService 인스턴스 생성
func NewOptimizationService() *OptimizationService {
	return &OptimizationService{
		pythonScriptPath: "./scripts/aggregator_optimization.py",
		inputDir:         "./temp/input",
		outputDir:        "./temp/output",
	}
}

// 집계자 배치 최적화 실행
func (s *OptimizationService) RunOptimization(request OptimizationRequest) (*OptimizationResponse, error) {
	startTime := time.Now()

	// 1. 임시 디렉토리 생성
	if err := s.createTempDirectories(); err != nil {
		return nil, fmt.Errorf("임시 디렉토리 생성 실패: %v", err)
	}

	// 2. 입력 데이터를 JSON 파일로 저장
	inputFilePath, err := s.saveInputData(request)
	if err != nil {
		return nil, fmt.Errorf("입력 데이터 저장 실패: %v", err)
	}

	// 3. Python 스크립트 실행
	outputFilePath := filepath.Join(s.outputDir, "optimization_result.json")
	if err := s.executePythonScript(inputFilePath, outputFilePath); err != nil {
		return nil, fmt.Errorf("Python 스크립트 실행 실패: %v", err)
	}

	// 4. 결과 파일 읽기
	result, err := s.readOptimizationResult(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("최적화 결과 읽기 실패: %v", err)
	}

	// 5. 실행 시간 계산
	executionTime := time.Since(startTime).Seconds()
	result.ExecutionTime = executionTime
	result.Status = "completed"

	// Python에서 에러가 발생한 경우 처리
	if result.Status == "error" {
		return nil, fmt.Errorf("Python 최적화 에러: %s", result.Message)
	}

	// 6. 임시 파일 정리
	s.cleanupTempFiles(inputFilePath, outputFilePath)

	return result, nil
}

// 하위 호환성을 위한 변환 함수
func (s *OptimizationService) RunOptimizationLegacy(request OptimizationRequest) (*OptimizationResponse, error) {
	// 새로운 형식으로 최적화 실행
	newResult, err := s.RunOptimization(request)
	if err != nil {
		return nil, err
	}

	// 기존 형식으로 변환 (필요한 경우)
	legacyResults := make([]OptimizationResult, len(newResult.OptimizedOptions))
	for i, option := range newResult.OptimizedOptions {
		legacyResults[i] = OptimizationResult{
			Rank:             option.Rank,
			Region:           option.Region,
			InstanceType:     option.InstanceType,
			EstimatedCost:    option.EstimatedMonthlyCost,
			EstimatedLatency: option.AvgLatency,
			CloudProvider:    option.CloudProvider,
		}
	}

	// 기존 구조체 형식으로 반환
	legacyResponse := &OptimizationResponse{
		Status:           newResult.Status,
		Summary:          newResult.Summary,
		OptimizedOptions: newResult.OptimizedOptions,
		Message:          newResult.Message,
		ExecutionTime:    newResult.ExecutionTime,
	}

	return legacyResponse, nil
}


// 임시 디렉토리 생성
func (s *OptimizationService) createTempDirectories() error {
	dirs := []string{s.inputDir, s.outputDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// 입력 데이터를 JSON 파일로 저장
func (s *OptimizationService) saveInputData(request OptimizationRequest) (string, error) {
	// 타임스탬프를 사용해 고유한 파일명 생성
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("optimization_input_%d.json", timestamp)
	filePath := filepath.Join(s.inputDir, filename)

	// JSON으로 직렬화
	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	// 파일로 저장
	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// Python 스크립트 실행
func (s *OptimizationService) executePythonScript(inputPath, outputPath string) error {
	// Python 스크립트 실행 명령어 구성
	cmd := exec.Command("python3", s.pythonScriptPath, inputPath, outputPath)
	
	// 환경변수 설정 (필요한 경우)
	cmd.Env = os.Environ()
	
	// 스크립트 실행
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Python 스크립트 실행 오류: %v, 출력: %s", err, string(output))
	}

	// 출력 파일이 생성되었는지 확인
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("Python 스크립트가 결과 파일을 생성하지 못했습니다: %s", outputPath)
	}

	return nil
}

// 최적화 결과 파일 읽기
func (s *OptimizationService) readOptimizationResult(outputPath string) (*OptimizationResponse, error) {
	// 파일 읽기
	jsonData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	// JSON 파싱
	var result OptimizationResponse
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// 임시 파일 정리
func (s *OptimizationService) cleanupTempFiles(filePaths ...string) {
	for _, path := range filePaths {
		if err := os.Remove(path); err != nil {
			// 로그 출력만 하고 에러는 무시 (정리 실패가 전체 프로세스를 중단시키지 않도록)
			fmt.Printf("임시 파일 삭제 실패: %s, 오류: %v\n", path, err)
		}
	}
}

// Python 스크립트 존재 여부 확인
func (s *OptimizationService) ValidatePythonScript() error {
	if _, err := os.Stat(s.pythonScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("Python 스크립트를 찾을 수 없습니다: %s", s.pythonScriptPath)
	}
	return nil
}

// Python 환경 확인
func (s *OptimizationService) ValidatePythonEnvironment() error {
	cmd := exec.Command("python3", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Python3이 설치되어 있지 않거나 PATH에 없습니다: %v", err)
	}
	return nil
}

// 의존성 패키지 확인
func (s *OptimizationService) ValidatePythonDependencies() error {
	requiredPackages := []string{"psycopg2", "deap", "numpy", "python-dotenv"}
	
	for _, pkg := range requiredPackages {
		cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", pkg))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("필수 Python 패키지가 설치되지 않았습니다: %s", pkg)
		}
	}
	
	return nil
}