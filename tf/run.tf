resource "google_service_account" "blog" {
  account_id = "blog-builder"

  depends_on = [module.enabled_google_apis]
}

resource "google_project_iam_member" "blog_act_as" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.blog.email}"
}

resource "google_project_iam_member" "blog_logs_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.blog.email}"
}

resource "google_artifact_registry_repository" "blog" {
  project       = var.project_id
  location      = "us"
  repository_id = "blog"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository_iam_member" "blog" {
  project    = google_artifact_registry_repository.blog.project
  location   = google_artifact_registry_repository.blog.location
  repository = google_artifact_registry_repository.blog.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.blog.email}"
}

resource "google_cloudbuild_trigger" "github" {
  project         = var.project_id
  location        = "us-central1"
  service_account = google_service_account.blog.id

  repository_event_config {
    repository = "projects/jameshartig-blog/locations/us-central1/connections/github-jameshartig/repositories/jameshartig-blog"
    push {
      branch = "^main$"
    }
  }

  build {
    images = [
      "${google_artifact_registry_repository.blog.registry_uri}/blog:latest",
    ]

    step {
      args = [
        "build",
        "-t",
        "${google_artifact_registry_repository.blog.registry_uri}/blog:latest",
        ".",
      ]
      name = "gcr.io/cloud-builders/docker"
    }
  }
}

resource "google_cloud_run_v2_service" "blog" {
  project  = var.project_id
  name     = "blog"
  location = "global"
  ingress  = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  #default_uri_disabled = true

  scaling {
    min_instance_count = 0
    max_instance_count = 1
  }

  multi_region_settings {
    regions = var.blog_run_regions
  }

  template {
    max_instance_request_concurrency = 1000
    service_account                  = google_service_account.blog.email

    containers {
      image = "${google_artifact_registry_repository.blog.registry_uri}/blog:latest"
      resources {
        limits = {
          cpu    = "2"
          memory = "1024Mi"
        }
        cpu_idle          = true
        startup_cpu_boost = true
      }
      # TODO: liveness_probe
      # TODO: startup_probe
    }

    vpc_access {
      network_interfaces {
        network = google_compute_network.default.id
      }
    }
  }
}

resource "google_compute_region_network_endpoint_group" "blog" {
  for_each = toset(var.blog_run_regions)

  project               = var.project_id
  name                  = "blog-${each.value}"
  network_endpoint_type = "SERVERLESS"
  region                = each.value

  cloud_run {
    service = google_cloud_run_v2_service.blog.name
  }
}
