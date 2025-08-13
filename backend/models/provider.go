package models

// Provider는 클라우드 제공업체 정보를 저장하는 구조체
type Provider struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"not null;unique;size:10;index"`
}

// 테이블명 설정
func (Provider) TableName() string {
	return "providers"
}
