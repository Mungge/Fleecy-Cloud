package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// getEncryptionKey 환경변수에서 암호화 키를 가져오거나 기본값 사용
func getEncryptionKey() []byte {
	key := os.Getenv("SSH_ENCRYPTION_KEY")
	if key == "" {
		// 개발용 기본 키 (프로덕션에서는 반드시 환경변수로 설정해야 함)
		key = "fleecy-cloud-ssh-key-encryption-secret-2024"
	}

	// SHA-256으로 32바이트 키 생성
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// EncryptPrivateKey Private Key를 AES-256-GCM으로 암호화
func EncryptPrivateKey(privateKey string) (string, error) {
	if privateKey == "" {
		return "", fmt.Errorf("private key is empty")
	}

	// 암호화 키 가져오기
	key := getEncryptionKey()

	// AES cipher 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// GCM 모드 사용
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// 랜덤 nonce 생성
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// 암호화
	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)

	// 문자열로 반환
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey 암호화된 Private Key를 복호화
func DecryptPrivateKey(encryptedPrivateKey string) (string, error) {
	if encryptedPrivateKey == "" {
		return "", fmt.Errorf("encrypted private key is empty")
	}

	// Base64 디코딩
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	// 암호화 키 가져오기
	key := getEncryptionKey()

	// AES cipher 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// GCM 모드 사용
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// nonce 크기 확인
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// nonce와 실제 암호화된 데이터 분리
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 복호화 수행
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}
