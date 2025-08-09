# versions.tf
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

# Provider 설정
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}