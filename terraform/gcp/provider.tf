terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  credentials = var.gcp_credentials_json  # JSON 내용 직접 사용
  project     = var.project_id
  region      = var.region
}