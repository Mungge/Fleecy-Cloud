# variables.tf
variable "project_id" {
  description = "GCP 프로젝트 ID"
  type        = string
}

variable "project_name" {
  description = "프로젝트 이름 (리소스 명명에 사용)"
  type        = string
}

variable "region" {
  description = "GCP 리전"
  type        = string
  default     = "asia-northeast3"  # 서울 리전
}

variable "zone" {
  description = "GCP 존"
  type        = string
  default     = "asia-northeast3-a"  # 서울 존
}

variable "environment" {
  description = "환경 (dev, prod)"
  type        = string
  default     = "dev"
}

variable "instance_type" {
  description = "VM 인스턴스 타입"
  type        = string
  default     = "f1-micro"  # AWS t2.micro과 유사
}

variable "allowed_ips" {
  description = "접근 허용할 IP CIDR 블록들"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "custom_ports" {
  description = "프로덕션 환경에서 허용할 사용자 지정 포트들"
  type        = list(number)
  default     = [8080, 9000]
}

variable "ssh_public_key_path" {
  description = "SSH 공개키 파일 경로"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "ssh_username" {
  description = "SSH 사용자명"
  type        = string
  default     = "ubuntu"
}