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
	"gorm.io/gorm"
)

// InitializeDataFromAssets는 asset 폴더의 CSV 파일들로부터 데이터를 초기화합니다
func InitializeDataFromAssets(db *gorm.DB) error {
	log.Println("Starting data initialization from asset files...")

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
		if len(record) < 7 {
			log.Printf("Skipping invalid record at line %d: insufficient columns", i+2)
			continue
		}

		vcpuCount, err := strconv.Atoi(record[3])
		if err != nil {
			log.Printf("Skipping record at line %d: invalid vcpu_count %s", i+2, record[3])
			continue
		}

		memoryGB, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Skipping record at line %d: invalid memory_gb %s", i+2, record[4])
			continue
		}

		onDemandPrice, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Printf("Skipping record at line %d: invalid on_demand_price %s", i+2, record[6])
			continue
		}

		cloudPrice := models.CloudPrice{
			CloudName:       strings.TrimSpace(record[0]),
			RegionName:      strings.TrimSpace(record[1]),
			InstanceType:    strings.TrimSpace(record[2]),
			VCPUCount:       vcpuCount,
			MemoryGB:        memoryGB,
			OperatingSystem: strings.TrimSpace(record[5]),
			OnDemandPrice:   onDemandPrice,
		}

		cloudPrices = append(cloudPrices, cloudPrice)
	}

	// 배치로 데이터 삽입
	if len(cloudPrices) > 0 {
		if err := db.CreateInBatches(cloudPrices, 1000).Error; err != nil {
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

		sourceProvider := strings.TrimSpace(record[1])
		sourceRegion := strings.TrimSpace(record[2])
		targetProvider := strings.TrimSpace(record[4])
		targetRegion := strings.TrimSpace(record[5])

		// 중복 확인용 키 생성
		pairKey := fmt.Sprintf("%s-%s-%s-%s", sourceProvider, sourceRegion, targetProvider, targetRegion)

		// 동일 방향의 기존 값이 있으면 더 좋은(작은) 지연시간으로 갱신
		if idx, exists := processedIdx[pairKey]; exists {
			if avgLatency < cloudLatencies[idx].AvgLatency {
				cloudLatencies[idx].AvgLatency = avgLatency
			}
			continue
		}

		cloudLatency := models.CloudLatency{
			SourceProvider: sourceProvider,
			SourceRegion:   sourceRegion,
			TargetProvider: targetProvider,
			TargetRegion:   targetRegion,
			AvgLatency:     avgLatency,
		}

	cloudLatencies = append(cloudLatencies, cloudLatency)
	processedIdx[pairKey] = len(cloudLatencies) - 1
	}

	// 배치로 데이터 삽입
	if len(cloudLatencies) > 0 {
		if err := db.CreateInBatches(cloudLatencies, 1000).Error; err != nil {
			return fmt.Errorf("failed to insert cloud latencies: %w", err)
		}
		log.Printf("Successfully inserted %d cloud latency records", len(cloudLatencies))
	} else {
		log.Println("No valid cloud latency records found to insert")
	}

	return nil
}
