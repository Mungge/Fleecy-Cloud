terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0" # AWS 프로바이더의 버전은 5.x를 사용
    }
  }
}

provider "aws" {
  region = var.aws_region
  access_key = local.aws_access_key
  secret_key = local.aws_secret_key
}