package models

// Region은 클라우드 리전 정보를 저장하는 구조체
type Region struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"not null;unique;size:30;index"`
}

// 테이블명 설정
func (Region) TableName() string {
	return "regions"
}
