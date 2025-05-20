variable "aws_region" {
  description = "AWS 리전"
  default     = "ap-northeast-2" # 서울 리전
}

variable "project_name" {
  description = "사용자 지정 프로젝트 이름"
  default     = "fleecy-cloud"
}

variable "environment" {
  description = "환경 (dev, staging, prod)"
  default     = "dev"
}