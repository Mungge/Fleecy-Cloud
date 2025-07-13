package models

import "fmt"

// CloudPrice는 클라우드 인스턴스 가격 정보를 저장하는 구조체
type CloudPrice struct {
	ID              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	CloudName       string    `json:"cloud_name" gorm:"not null;index;size:20"`
	RegionName      string    `json:"region_name" gorm:"not null;index;size:30"`
	InstanceType    string    `json:"instance_type" gorm:"not null;index;size:30"`
	VCPUCount       int       `json:"vcpu_count" gorm:"not null;index"`
	MemoryGB        int       `json:"memory_gb" gorm:"not null;index"`
	OperatingSystem string    `json:"operating_system" gorm:"default:'Linux';size:50"`
	OnDemandPrice   float64   `json:"on_demand_price" gorm:"not null;type:decimal(10,6)"`

	//관계 설정

}

// 테이블명 설정
func (CloudPrice) TableName() string {
	return "cloud_prices"
}

func (cp *CloudPrice) String() string {
	return fmt.Sprintf("%s %s %s: $%.4f/hour (%dvCPU, %dGB)", 
		cp.CloudName, cp.RegionName, cp.InstanceType, 
		cp.OnDemandPrice, cp.VCPUCount, cp.MemoryGB)
}

func (cp *CloudPrice) HourlyRate() float64 {
	return cp.OnDemandPrice
}

func (cp *CloudPrice) DailyRate() float64 {
	return cp.OnDemandPrice * 24
}

func (cp *CloudPrice) MonthlyRate() float64 {
	return cp.DailyRate() * 30
}

func (cp *CloudPrice) AnnualRate() float64 {
	return cp.MonthlyRate() * 12
}

// 하루 이용 기준 vCPU당 비용 계산
func (cp *CloudPrice) PricePerVCPU() float64 {
	if cp.VCPUCount == 0 {
		return 0
	}
	return cp.DailyRate() / float64(cp.VCPUCount)
}

// 하루 이용 기준 GB당 메모리 비용 계산
func (cp *CloudPrice) PricePerGB() float64 {
	if cp.MemoryGB == 0 {
		return 0
	}
	return cp.DailyRate() / float64(cp.MemoryGB)
}


