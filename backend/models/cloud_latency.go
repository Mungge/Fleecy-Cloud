package models

// CloudLatency는 리전 간의 지연 시간 측정 결과를 저장하는 구조체
type CloudLatency struct {
	ID             int        `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceProvider string     `json:"source_provider" gorm:"not null;size:30;index"`
	SourceRegion   string     `json:"source_region" gorm:"not null;size:30;index"`
	TargetProvider string     `json:"target_provider" gorm:"not null;size:30;index"`
	TargetRegion   string     `json:"target_region" gorm:"not null;size:30;index"`
	AvgLatency     float64    `json:"avg_latency" gorm:"type:decimal(8,2);index"`
}

// 테이블명 설정
func (CloudLatency) TableName() string {
	return "cloud_latency"
}


