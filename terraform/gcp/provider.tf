terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  credentials = local.gcp_credentials  # JSON 내용 직접 사용
  project     = var.project_id
  region      = var.region
}