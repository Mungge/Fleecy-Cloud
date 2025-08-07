terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  credentials = local.gcp_credentials  # 프로덕션: 변수값, 개발: 파일
  project     = var.project_id
  region      = var.region
}