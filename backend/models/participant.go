package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Participant struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    int64     `json:"user_id" gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name      string    `json:"name" gorm:"not null;type:varchar(255)"`
	Status    string    `json:"status" gorm:"type:varchar(50);default:'inactive';check:status IN ('active','inactive')"`
	Metadata  string    `json:"metadata,omitempty" gorm:"type:text"`

	// OpenStack 클라우드 관련 필드
	OpenStackEndpoint    string `json:"openstack_endpoint,omitempty" gorm:"type:varchar(500)"`     // OpenStack 인증 엔드포인트
	OpenStackRegion      string `json:"openstack_region,omitempty" gorm:"type:varchar(100)"`       // OpenStack 리전
	OpenStackApplicationCredentialID     string `json:"openstack_app_credential_id,omitempty" gorm:"type:varchar(255)"`     // Application Credential ID
	OpenStackApplicationCredentialSecret string `json:"openstack_app_credential_secret,omitempty" gorm:"type:varchar(500)"` // Application Credential Secret -> 암호화 필요

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 관계 설정
	User                 User                   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FederatedLearnings   []FederatedLearning    `json:"federated_learnings,omitempty" gorm:"many2many:participant_federated_learnings;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	VirtualMachines      []VirtualMachine       `json:"virtual_machines,omitempty" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Participant) TableName() string {
	return "participants"
}

// GetAvailableVMs는 사용 가능한 VM 목록을 반환합니다
func (p *Participant) GetAvailableVMs() []VirtualMachine {
	var availableVMs []VirtualMachine
	for _, vm := range p.VirtualMachines {
		if vm.IsAvailable() {
			availableVMs = append(availableVMs, vm)
		}
	}
	return availableVMs
}

// GetBusyVMs는 작업 중인 VM 목록을 반환합니다
func (p *Participant) GetBusyVMs() []VirtualMachine {
	var busyVMs []VirtualMachine
	for _, vm := range p.VirtualMachines {
		if vm.IsBusy() {
			busyVMs = append(busyVMs, vm)
		}
	}
	return busyVMs
}

// GetTotalCapacity는 전체 VM 수용량을 계산합니다
func (p *Participant) GetTotalCapacity() int {
	return len(p.VirtualMachines)
}

// GetAvailableCapacity는 사용 가능한 VM 수를 반환합니다
func (p *Participant) GetAvailableCapacity() int {
	return len(p.GetAvailableVMs())
}

// GORM Hooks - BeforeCreate는 레코드 생성 전에 호출됩니다
func (p *Participant) BeforeCreate(tx *gorm.DB) error {
	// ID가 없으면 UUID 생성
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	// 민감한 정보 암호화
	if err := p.encryptSensitiveData(); err != nil {
		return fmt.Errorf("민감한 데이터 암호화 실패: %v", err)
	}

	return nil
}

// BeforeUpdate는 레코드 업데이트 전에 호출됩니다
func (p *Participant) BeforeUpdate(tx *gorm.DB) error {
	// 민감한 정보 암호화
	if err := p.encryptSensitiveData(); err != nil {
		return fmt.Errorf("민감한 데이터 암호화 실패: %v", err)
	}

	return nil
}

// AfterFind는 레코드 조회 후에 호출됩니다
func (p *Participant) AfterFind(tx *gorm.DB) error {
	// 민감한 정보 복호화
	if err := p.decryptSensitiveData(); err != nil {
		return fmt.Errorf("민감한 데이터 복호화 실패: %v", err)
	}

	return nil
}

// 민감한 데이터 암호화 (실제 구현에서는 더 안전한 키 관리 필요)
func (p *Participant) encryptSensitiveData() error {
	key := []byte("32-byte-long-key-for-encryption!") // 실제로는 환경변수나 키 관리 시스템에서 가져와야 함

	if p.OpenStackApplicationCredentialSecret != "" {
		encrypted, err := encrypt(p.OpenStackApplicationCredentialSecret, key)
		if err != nil {
			return err
		}
		p.OpenStackApplicationCredentialSecret = encrypted
	}

	return nil
}

// 민감한 데이터 복호화
func (p *Participant) decryptSensitiveData() error {
	key := []byte("32-byte-long-key-for-encryption!") // 실제로는 환경변수나 키 관리 시스템에서 가져와야 함

	if p.OpenStackApplicationCredentialSecret != "" && len(p.OpenStackApplicationCredentialSecret) > 20 {
		decrypted, err := decrypt(p.OpenStackApplicationCredentialSecret, key)
		if err != nil {
			return err
		}
		p.OpenStackApplicationCredentialSecret = decrypted
	}

	return nil
}

// 간단한 AES 암호화 함수
func encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 간단한 AES 복호화 함수
func decrypt(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("암호화된 데이터가 너무 짧습니다")
	}

	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (p *Participant) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("참여자 이름은 필수입니다")
	}

	if p.OpenStackEndpoint == "" {
		return fmt.Errorf("OpenStack 엔드포인트는 필수입니다")
	}

	hasAppCredential := p.OpenStackApplicationCredentialID != "" && p.OpenStackApplicationCredentialSecret != ""

	if !hasAppCredential {
		return fmt.Errorf("application Credential 인증 정보가 필요합니다")
	}

	return nil
}

// 상태 관리 메서드들
func (p *Participant) IsActive() bool {
	return p.Status == "active"
}

func (p *Participant) IsInactive() bool {
	return p.Status == "inactive"
}

func (p *Participant) SetActive() {
	p.Status = "active"
	p.UpdatedAt = time.Now()
}

func (p *Participant) SetInactive() {
	p.Status = "inactive"
	p.UpdatedAt = time.Now()
}

// 인증 방식 확인 메서드들
func (p *Participant) HasApplicationCredential() bool {
	return p.OpenStackApplicationCredentialID != "" && p.OpenStackApplicationCredentialSecret != ""
}
