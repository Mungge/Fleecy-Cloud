package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InitializeDataFromAssets는 asset 폴더의 CSV 파일들로부터 데이터를 초기화합니다
func InitializeDataFromAssets(db *gorm.DB) error {
	log.Println("Starting data initialization from asset files...")

	// Provider와 Region 초기화
	if err := initializeProvidersAndRegions(db); err != nil {
		return fmt.Errorf("failed to initialize providers and regions: %w", err)
	}

	// Cloud Price 데이터 초기화
	if err := initializeCloudPrices(db); err != nil {
		return fmt.Errorf("failed to initialize cloud prices: %w", err)
	}

	// Cloud Latency 데이터 초기화
	if err := initializeCloudLatencies(db); err != nil {
		return fmt.Errorf("failed to initialize cloud latencies: %w", err)
	}

	log.Println("Data initialization completed successfully!")
	return nil
}

// findAssetPath는 asset 폴더의 경로를 찾습니다
func findAssetPath(filename string) string {
	// 현재 디렉토리에서 asset 폴더 확인
	assetPaths := []string{
		filepath.Join("asset", filename),
		filepath.Join("..", "asset", filename),
		filepath.Join("..", "..", "asset", filename),
	}

	for _, path := range assetPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 기본값 반환
	return filepath.Join("asset", filename)
}

// initializeProvidersAndRegions는 providers와 regions 기본 데이터를 초기화합니다
func initializeProvidersAndRegions(db *gorm.DB) error {
	log.Println("Initializing providers and regions...")

	providerRepo := repository.NewProviderRepository(db)
	regionRepo := repository.NewRegionRepository(db)

	// Provider 데이터 초기화
	providers := []struct {
		Name string
	}{
		{"aws"},
		{"gcp"},
	}

	for _, p := range providers {
		if _, err := providerRepo.CreateOrGetProvider(p.Name); err != nil {
			return fmt.Errorf("failed to create provider %s: %w", p.Name, err)
		}
	}

	// Region 데이터 초기화 (리전 목록)
	regions := []string{
		"sa-east-1",
		"asia-east2",
		"ap-south-1",
		"eu-west-1",
		"europe-central2",
		"eu-south-2",
		"ap-southeast-1",
		"eu-central-1",
		"af-south-1",
		"europe-west12",
		"me-south-1",
		"ca-central-1",
		"northamerica-northeast1",
		"ap-northeast-2",
		"asia-southeast1",
		"us-west-1",
		"us-east-1",
		"europe-west4",
		"asia-south1",
		"australia-southeast1",
		"southamerica-east1",
		"asia-northeast1",
		"us-west1",
		"us-east4",
		"ap-northeast-1",
		"asia-northeast3",
	}

	for _, regionName := range regions {
		if _, err := regionRepo.CreateOrGetRegion(regionName); err != nil {
			return fmt.Errorf("failed to create region %s: %w", regionName, err)
		}
	}

	log.Println("Providers and regions initialized successfully")
	return nil
}

// initializeCloudPrices는 cloud_price_AWS.csv 파일로부터 가격 데이터를 초기화합니다
func initializeCloudPrices(db *gorm.DB) error {
	// 이미 데이터가 있는지 확인
	var count int64
	if err := db.Model(&models.CloudPrice{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count existing cloud prices: %w", err)
	}

	if count > 0 {
		log.Printf("Cloud price data already exists (%d records), skipping initialization", count)
		return nil
	}

	log.Println("Initializing cloud price data...")

	providerRepo := repository.NewProviderRepository(db)
	regionRepo := repository.NewRegionRepository(db)

	// CSV 파일 읽기
	cloudPriceFile := findAssetPath("cloud_price_AWS.csv")
	file, err := os.Open(cloudPriceFile)
	if err != nil {
		return fmt.Errorf("failed to open cloud_price_AWS.csv at %s: %w", cloudPriceFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	// 헤더 제거
	if len(records) > 0 {
		records = records[1:]
	}

	var cloudPrices []models.CloudPrice
	for i, record := range records {
		if len(record) < 6 {
			log.Printf("Skipping invalid record at line %d: insufficient columns", i+2)
			continue
		}

		vcpuCount, err := strconv.Atoi(record[3])
		if err != nil {
			log.Printf("Skipping record at line %d: invalid v_cpu_count %s", i+2, record[3])
			continue
		}

		memoryGB, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Skipping record at line %d: invalid memory_gb %s", i+2, record[4])
			continue
		}

		onDemandPrice, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Skipping record at line %d: invalid on_demand_price %s", i+2, record[5])
			continue
		}

		// Provider 조회
		cloudName := strings.TrimSpace(record[0])
		var providerName string
		switch strings.ToUpper(cloudName) {
		case "AWS":
			providerName = "aws"
		case "GCP":
			providerName = "gcp"
		default:
			providerName = strings.ToLower(cloudName)
		}
		
		provider, err := providerRepo.GetProviderByName(providerName)
		if err != nil {
			log.Printf("Skipping record at line %d: provider %s not found: %v", i+2, providerName, err)
			continue
		}

		// Region 조회
		regionName := strings.TrimSpace(record[1])
		region, err := regionRepo.GetRegionByName(regionName)
		if err != nil {
			log.Printf("Skipping record at line %d: region %s not found: %v", i+2, regionName, err)
			continue
		}

		cloudPrice := models.CloudPrice{
			ProviderID:      provider.ID,
			RegionID:        region.ID,
			InstanceType:    strings.TrimSpace(record[2]),
			VCPUCount:       vcpuCount,
			MemoryGB:        memoryGB,
			OnDemandPrice:   onDemandPrice,
		}

		cloudPrices = append(cloudPrices, cloudPrice)
	}

	// 배치로 데이터 삽입
	if len(cloudPrices) > 0 {
		// 복합 유니크 인덱스(provider_id, region_id, instance_type) 충돌 시 무시
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "provider_id"}, {Name: "region_id"}, {Name: "instance_type"}},
			DoNothing: true,
		}).CreateInBatches(cloudPrices, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert cloud prices: %w", err)
		}
		log.Printf("Successfully inserted %d cloud price records", len(cloudPrices))
	}

	return nil
}

// initializeCloudLatencies는 latency_results.csv 파일로부터 지연시간 데이터를 초기화합니다
func initializeCloudLatencies(db *gorm.DB) error {
	// 이미 데이터가 있는지 확인
	var count int64
	if err := db.Model(&models.CloudLatency{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count existing cloud latencies: %w", err)
	}

	if count > 0 {
		log.Printf("Cloud latency data already exists (%d records), skipping initialization", count)
		return nil
	}

	log.Println("Initializing cloud latency data...")

	providerRepo := repository.NewProviderRepository(db)
	regionRepo := repository.NewRegionRepository(db)

	// CSV 파일 읽기
	latencyFile := findAssetPath("latency_results.csv")
	file, err := os.Open(latencyFile)
	if err != nil {
		return fmt.Errorf("failed to open latency_results.csv at %s: %w", latencyFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	// 헤더 제거
	if len(records) > 0 {
		records = records[1:]
	}

	cloudLatencies := make([]models.CloudLatency, 0)
	processedIdx := make(map[string]int)

	for i, record := range records {
		if len(record) < 9 {
			log.Printf("Skipping invalid record at line %d: insufficient columns", i+2)
			continue
		}

		// success가 TRUE인 경우만 처리
		if strings.ToUpper(strings.TrimSpace(record[7])) != "TRUE" {
			continue
		}

		avgLatency, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			log.Printf("Skipping record at line %d: invalid avg_latency %s", i+2, record[8])
			continue
		}

		sourceProviderName := strings.TrimSpace(record[1])
		var sourceProviderName2 string
		switch strings.ToUpper(sourceProviderName) {
		case "AWS":
			sourceProviderName2 = "aws"
		case "GCP":
			sourceProviderName2 = "gcp"
		default:
			sourceProviderName2 = strings.ToLower(sourceProviderName)
		}
		
		sourceRegionName := strings.TrimSpace(record[2])
		
		targetProviderName := strings.TrimSpace(record[4])
		var targetProviderName2 string
		switch strings.ToUpper(targetProviderName) {
		case "AWS":
			targetProviderName2 = "aws"
		case "GCP":
			targetProviderName2 = "gcp"
		default:
			targetProviderName2 = strings.ToLower(targetProviderName)
		}
		targetRegionName := strings.TrimSpace(record[5])

		// Provider 조회
		sourceProvider, err := providerRepo.GetProviderByName(sourceProviderName2)
		if err != nil {
			log.Printf("Skipping record at line %d: source provider %s not found: %v", i+2, sourceProviderName2, err)
			continue
		}

		targetProvider, err := providerRepo.GetProviderByName(targetProviderName2)
		if err != nil {
			log.Printf("Skipping record at line %d: target provider %s not found: %v", i+2, targetProviderName2, err)
			continue
		}

		// Region 조회
		sourceRegion, err := regionRepo.GetRegionByName(sourceRegionName)
		if err != nil {
			log.Printf("Skipping record at line %d: source region %s not found: %v", i+2, sourceRegionName, err)
			continue
		}

		targetRegion, err := regionRepo.GetRegionByName(targetRegionName)
		if err != nil {
			log.Printf("Skipping record at line %d: target region %s not found: %v", i+2, targetRegionName, err)
			continue
		}

		// 중복 확인용 키 생성
		pairKey := fmt.Sprintf("%d-%d-%d-%d", sourceProvider.ID, sourceRegion.ID, targetProvider.ID, targetRegion.ID)

		// 동일 방향의 기존 값이 있으면 더 좋은(작은) 지연시간으로 갱신
		if idx, exists := processedIdx[pairKey]; exists {
			if avgLatency < cloudLatencies[idx].AvgLatency {
				cloudLatencies[idx].AvgLatency = avgLatency
			}
			continue
		}

		cloudLatency := models.CloudLatency{
			SourceProviderID: sourceProvider.ID,
			SourceRegionID:   sourceRegion.ID,
			TargetProviderID: targetProvider.ID,
			TargetRegionID:   targetRegion.ID,
			AvgLatency:       avgLatency,
		}

		cloudLatencies = append(cloudLatencies, cloudLatency)
		processedIdx[pairKey] = len(cloudLatencies) - 1
	}

	// 배치로 데이터 삽입
	if len(cloudLatencies) > 0 {
		// 복합 유니크 인덱스(source_provider_id, source_region_id, target_provider_id, target_region_id) 충돌 시 무시
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "source_provider_id"}, {Name: "source_region_id"}, {Name: "target_provider_id"}, {Name: "target_region_id"}},
			DoNothing: true,
		}).CreateInBatches(cloudLatencies, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert cloud latencies: %w", err)
		}
		log.Printf("Successfully inserted %d cloud latency records", len(cloudLatencies))
	} else {
		log.Println("No valid cloud latency records found to insert")
	}

	return nil
}
