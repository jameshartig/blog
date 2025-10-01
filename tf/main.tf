terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.3.0"
    }
  }
}

provider "google" {
  project = var.project_id
}


variable "project_id" {
  description = "gcp project ID"
  default     = "jameshartig-blog"
}

variable "networks" {
  default = {
    "us-central1": "10.0.0.0/24",
    "europe-west1": "10.0.16.0/24",
  }
}

variable "blog_run_regions" {
  default = [
    "us-central1",
    "europe-west1",
  ]
}

module "enabled_google_apis" {
  source  = "terraform-google-modules/project-factory/google//modules/project_services"
  version = "~> 18.1"

  project_id                  = var.project_id
  disable_services_on_destroy = false

  activate_apis = [
    "run.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "cloudtrace.googleapis.com",
    "cloudbuild.googleapis.com",
    "storage.googleapis.com",
    "containerregistry.googleapis.com",
    "containeranalysis.googleapis.com",
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "certificatemanager.googleapis.com",
    "developerconnect.googleapis.com"
  ]
}
