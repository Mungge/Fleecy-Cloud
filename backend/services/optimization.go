package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// 최적화 요청 구조체
type OptimizationRequest struct {
	FederatedLearning struct {
		Name          string        `json:"name"`
		Description   string        `json:"description"`
		ModelType     string        `json:"modelType"`
		Algorithm     string        `json:"algorithm"`
		Rounds        int           `json:"rounds"`
		Participants  []Participant `json:"participants"`
		ModelFileName *string       `json:"modelFileName,omitempty"`
	} `json:"federatedLearning"`
	AggregatorConfig struct {
		MaxBudget  int `json:"maxBudget"`
		MaxLatency int `json:"maxLatency"`
		WeightBalance *int `json:"weightBalance,omitempty"`
	} `json:"aggregatorConfig"`
}

// 참여자 구조체
type Participant struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	Region            string `json:"region"`
	OpenstackEndpoint string `json:"openstack_endpoint"`
}

// 최적화 응답 구조체
type OptimizationResponse struct {
	Status           string              `json:"status"`
	Summary          OptimizationSummary `json:"summary"`
	OptimizedOptions []AggregatorOption  `json:"optimizedOptions"`
	Message          string              `json:"message"`
	ExecutionTime    float64             `json:"executionTime,omitempty"` // Go에서 추가
}

// 최적화 요약 정보
type OptimizationSummary struct {
	TotalParticipants     int         `json:"totalParticipants"`
	ParticipantRegions    []string    `json:"participantRegions"`
	TotalCandidateOptions int         `json:"totalCandidateOptions"`
	FeasibleOptions       int         `json:"feasibleOptions"`
	Constraints           interface{} `json:"constraints"`
	ModelInfo             interface{} `json:"modelInfo"`
}

// 집계자 옵션 (새로운 구조)
type AggregatorOption struct {
	Rank                 int     `json:"rank"`
	Region               string  `json:"region"`
	InstanceType         string  `json:"instanceType"`
	CloudProvider        string  `json:"cloudProvider"`
	EstimatedMonthlyCost float64 `json:"estimatedMonthlyCost"`
	EstimatedHourlyPrice float64 `json:"estimatedHourlyPrice"`
	AvgLatency           float64 `json:"avgLatency"`
	MaxLatency           float64 `json:"maxLatency"`
	VCPU                 int     `json:"vcpu"`
	Memory               int     `json:"memory"`
	RecommendationScore  float64 `json:"recommendationScore"`
}

// 최적화 결과 구조체
type OptimizationResult struct {
	Rank             int     `json:"rank"`
	Region           string  `json:"region"`
	InstanceType     string  `json:"instanceType"`
	EstimatedCost    float64 `json:"estimatedCost"`
	EstimatedLatency float64 `json:"estimatedLatency"`
	CloudProvider    string  `json:"cloudProvider"`
}

// OptimizationService 구조체
type OptimizationService struct {
	pythonScriptPath string
	tempBaseDir      string
	inputDir         string // 각 실행마다 runTemp/input 으로 설정
	outputDir        string // 각 실행마다 runTemp/output 으로 설정
}

// 새로운 OptimizationService 인스턴스 생성
func NewOptimizationService() *OptimizationService {
	return &OptimizationService{
		pythonScriptPath: "./scripts/aggregator_optimization.py",
		tempBaseDir:      "./temp",
	}
}

// 집계자 배치 최적화 실행
func (s *OptimizationService) RunOptimization(request OptimizationRequest) (*OptimizationResponse, error) {
	startTime := time.Now()

	if err := os.MkdirAll(s.tempBaseDir, 0o755); err != nil { // ✅ 베이스 temp 보장
		return nil, fmt.Errorf("temp 베이스 디렉토리 생성 실패: %w", err)
	}

	runTemp, err := os.MkdirTemp(s.tempBaseDir, "run-")
	if err != nil {
		return nil, fmt.Errorf("임시 디렉토리 생성 실패: %w", err)
	}
	// 항상 정리
	defer func() {
		if err := os.RemoveAll(runTemp); err != nil {
			fmt.Printf("임시 디렉토리 삭제 실패: %s, 오류: %v\n", runTemp, err)
		}

		if err := removeIfEmpty(s.tempBaseDir); err != nil {
		}
	}()

	s.inputDir = filepath.Join(runTemp, "input")
	s.outputDir = filepath.Join(runTemp, "output")

	if err := os.MkdirAll(s.inputDir, 0o755); err != nil {
		return nil, fmt.Errorf("input 디렉토리 생성 실패: %w", err)
	}
	if err := os.MkdirAll(s.outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("output 디렉토리 생성 실패: %w", err)
	}

	// 2. 입력 데이터를 JSON 파일로 저장
	inputFilePath, err := s.saveInputData(request)
	if err != nil {
		return nil, fmt.Errorf("입력 데이터 저장 실패: %v", err)
	}

	//3. Python 스크립트 실행
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
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// Python 스크립트 실행
func (s *OptimizationService) executePythonScript(inputPath, outputPath string) error {
	cmd := exec.Command("bash", "./scripts/run_optimizer.sh", inputPath, outputPath)
	cmd.Env = os.Environ() // .env는 파이썬 쪽에서 python-dotenv가 읽습니다.
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("쉘 스크립트 실행 오류: %v, 출력: %s", err, string(out))
	}
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("결과 파일이 생성되지 않았습니다: %s", outputPath)
	}
	return nil
}

// 최적화 결과 파일 읽기
func (s *OptimizationService) readOptimizationResult(outputPath string) (*OptimizationResponse, error) {
	// 파일 읽기
	jsonData, err := os.ReadFile(outputPath)
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

func removeIfEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err // 존재하지 않음 등은 호출부에서 무시해도 됨
	}
	if len(entries) > 0 {
		return fmt.Errorf("not empty")
	}
	return os.Remove(dir) // 비어 있을 때만 삭제됨
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
