package models

// CloudLatency는 리전 간의 지연 시간 측정 결과를 저장하는 구조체
type CloudLatency struct {
	ID               int     `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceProviderID int     `json:"source_provider_id" gorm:"not null;index;uniqueIndex:idx_latency_unique"`
	SourceRegionID   int     `json:"source_region_id" gorm:"not null;index;uniqueIndex:idx_latency_unique"`
	TargetProviderID int     `json:"target_provider_id" gorm:"not null;index;uniqueIndex:idx_latency_unique"`
	TargetRegionID   int     `json:"target_region_id" gorm:"not null;index;uniqueIndex:idx_latency_unique"`
	AvgLatency       float64  `json:"avg_latency" gorm:"type:decimal(8,2);index"`
	MinLatency       *float64 `json:"min_latency" gorm:"type:decimal(8,2);index"`
	MaxLatency       *float64 `json:"max_latency" gorm:"type:decimal(8,2);index"`

	// 관계 설정
	SourceProvider Provider `json:"source_provider" gorm:"foreignKey:SourceProviderID"`
	SourceRegion   Region   `json:"source_region" gorm:"foreignKey:SourceRegionID"`
	TargetProvider Provider `json:"target_provider" gorm:"foreignKey:TargetProviderID"`
	TargetRegion   Region   `json:"target_region" gorm:"foreignKey:TargetRegionID"`
}

// 테이블명 설정
func (CloudLatency) TableName() string {
	return "cloud_latency"
}


