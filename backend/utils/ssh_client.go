package utils

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient SSH 클라이언트 구조체
type SSHClient struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
}

// NewSSHClient SSH 클라이언트 생성
func NewSSHClient(host, port, user, privateKey string) *SSHClient {
	return &SSHClient{
		Host:       host,
		Port:       port,
		User:       user,
		PrivateKey: privateKey,
	}
}

// Connect SSH 연결 생성
func (c *SSHClient) Connect() (*ssh.Client, error) {
	// Private key 파싱
	key, err := ssh.ParsePrivateKey([]byte(c.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("private key 파싱 실패: %v", err)
	}

	// SSH 클라이언트 설정
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 보안상 주의 - 실제 환경에서는 적절한 호스트 키 검증 필요
		Timeout:         30 * time.Second,
	}

	// SSH 연결
	address := net.JoinHostPort(c.Host, c.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("SSH 연결 실패: %v", err)
	}

	return client, nil
}

// ExecuteCommand SSH를 통해 원격 명령(command) 실행
func (c *SSHClient) ExecuteCommand(command string) (string, string, error) {
	client, err := c.Connect()
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	// 세션 생성
	session, err := client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("세션 생성 실패: %v", err)
	}
	defer session.Close()

	// 명령 실행
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	return stdout.String(), stderr.String(), err
}

// UploadFile SSH를 통해 파일 업로드
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
	client, err := c.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	// SFTP 세션 생성
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("SFTP 세션 생성 실패: %v", err)
	}
	defer sftpClient.Close()

	// 로컬 파일 읽기
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("로컬 파일 읽기 실패: %v", err)
	}

	// 원격 파일 생성 및 쓰기
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("원격 파일 생성 실패: %v", err)
	}
	defer remoteFile.Close()

	_, err = remoteFile.Write(data)
	if err != nil {
		return fmt.Errorf("원격 파일 쓰기 실패: %v", err)
	}

	return nil
}

// UploadFileContent SSH를 통해 파일 내용 업로드
func (c *SSHClient) UploadFileContent(content, remotePath string) error {
	client, err := c.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	// SCP를 통한 파일 업로드 (간단한 방법)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("세션 생성 실패: %v", err)
	}
	defer session.Close()

	// SCP 명령을 통해 파일 내용 업로드
	cmd := fmt.Sprintf("cat > %s", remotePath)
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin 파이프 생성 실패: %v", err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write([]byte(content))
	}()

	return session.Run(cmd)
}

// CheckConnection SSH 연결 테스트
func (c *SSHClient) CheckConnection() error {
	_, _, err := c.ExecuteCommand("echo 'SSH connection test'")
	return err
}
