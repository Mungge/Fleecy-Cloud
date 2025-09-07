variable "aws_region" {
  description = "AWS 리전"
  type        = string
  default     = "ap-northeast-2"
}

variable "availability_zone" {
  description = "가용 영역"
  type        = string
  default     = ""
}

variable "project_name" {
  description = "프로젝트 이름"
  type        = string
  default     = "fleecy-cloud"
}

variable "instance_type" {
  description = "EC2 인스턴스 스펙"
  type        = string
  default     = "t2.micro"
}

variable "ssh_public_key_content" {
  description = "SSH 공개키 내용 (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
}

variable "ssh_username" {
  description = "SSH 사용자명"
  type        = string
  default     = "ubuntu"
}

variable "environment" {
  description = "환경 (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "allowed_ips" {
  description = "허용할 IP 주소 목록"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "custom_ports" {
  description = "추가로 열고 싶은 포트 목록 (개발환경이 아닐 때)"
  type        = list(number)
  default     = [8080, 9000, 5000, 9092]
}

variable "aws_access_key" {
  description = "AWS Access Key ID (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS Secret Access Key (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aggregator_id" {
  description = "집계자 고유 ID"
  type        = string
}

variable "startup_script" {
  description = "Startup script content"
  type        = string
}