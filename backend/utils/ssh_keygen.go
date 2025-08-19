package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// SSHKeyPair SSH 키 쌍 구조체
type SSHKeyPair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateSSHKeyPair RSA SSH 키 쌍을 생성합니다
func GenerateSSHKeyPair() (*SSHKeyPair, error) {
	// RSA 개인키 생성 (2048 비트)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %v", err)
	}

	// 개인키를 PEM 형식으로 인코딩
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// 공개키를 SSH 형식으로 변환
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %v", err)
	}

	// SSH 공개키를 문자열로 변환 (ssh-rsa 형식)
	publicKeyString := string(ssh.MarshalAuthorizedKey(publicKey))

	return &SSHKeyPair{
		PublicKey:  publicKeyString,
		PrivateKey: string(privateKeyBytes),
	}, nil
}
