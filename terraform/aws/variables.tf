variable "aws_region" {
  description = "AWS 리전"
  type        = string
  default     = "ap-northeast-2"
}

variable "availability_zone" {
  description = "가용 영역"
  type        = string
  default     = "ap-northeast-2a"
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

variable "key_pair_name" {
  description = "기존 키페어 이름 (미리 생성된 것)"
  type        = string
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
  default     = [8080, 9000, 5000]
}