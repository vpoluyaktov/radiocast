provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "storage.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudscheduler.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com"
  ])
  
  project = var.project_id
  service = each.value
  
  disable_dependent_services = false
  disable_on_destroy        = false
}

# GCS bucket for reports
resource "google_storage_bucket" "reports" {
  name     = var.reports_bucket_name
  location = var.region
  project  = var.project_id
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = var.environment == "production"
  }
  
  lifecycle_rule {
    condition {
      age = var.reports_retention_days
    }
    action {
      type = "Delete"
    }
  }
  
  dynamic "lifecycle_rule" {
    for_each = var.environment == "production" ? [1] : []
    content {
      condition {
        age                = 30
        with_state         = "NONCURRENT"
        num_newer_versions = 3
      }
      action {
        type = "Delete"
      }
    }
  }
  
  depends_on = [google_project_service.apis]
}

# Make reports bucket publicly readable
resource "google_storage_bucket_iam_member" "reports_public" {
  bucket = google_storage_bucket.reports.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Secret for OpenAI API key
resource "google_secret_manager_secret" "openai_api_key" {
  secret_id = "openai-api-key"
  project   = var.project_id
  
  replication {
    auto {}
  }
  
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "openai_api_key" {
  secret      = google_secret_manager_secret.openai_api_key.id
  secret_data = var.openai_api_key
}

# Service account for Cloud Run
resource "google_service_account" "radiocast" {
  account_id   = "radiocast-${var.environment}"
  display_name = "Radiocast Service Account (${title(var.environment)})"
  project      = var.project_id
}

# IAM permissions for service account
resource "google_project_iam_member" "radiocast_storage" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.radiocast.email}"
}

resource "google_project_iam_member" "radiocast_secrets" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.radiocast.email}"
}

resource "google_project_iam_member" "radiocast_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.radiocast.email}"
}

resource "google_project_iam_member" "radiocast_monitoring" {
  count   = var.enable_monitoring ? 1 : 0
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.radiocast.email}"
}

# Cloud Run service
resource "google_cloud_run_v2_service" "radiocast" {
  name     = var.service_name
  location = var.region
  project  = var.project_id
  
  template {
    service_account = google_service_account.radiocast.email
    
    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }
    
    containers {
      image = "gcr.io/${var.project_id}/radiocast:latest"
      
      ports {
        container_port = 8080
      }
      
      env {
        name  = "PORT"
        value = "8080"
      }
      
      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }
      
      env {
        name  = "GCS_BUCKET"
        value = google_storage_bucket.reports.name
      }
      
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }
      
      env {
        name  = "LOG_LEVEL"
        value = var.environment == "production" ? "info" : "debug"
      }
      
      env {
        name = "OPENAI_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.openai_api_key.secret_id
            version = "latest"
          }
        }
      }
      
      resources {
        limits = {
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }
        cpu_idle = var.environment != "production"
      }
      
      dynamic "startup_probe" {
        for_each = var.environment == "production" ? [1] : []
        content {
          http_get {
            path = "/health"
            port = 8080
          }
          initial_delay_seconds = 10
          timeout_seconds       = 5
          period_seconds        = 10
          failure_threshold     = 3
        }
      }
      
      dynamic "liveness_probe" {
        for_each = var.environment == "production" ? [1] : []
        content {
          http_get {
            path = "/health"
            port = 8080
          }
          initial_delay_seconds = 30
          timeout_seconds       = 5
          period_seconds        = 30
          failure_threshold     = 3
        }
      }
    }
    
    timeout = var.timeout
  }
  
  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
  
  depends_on = [google_project_service.apis]
}

# Allow unauthenticated access to Cloud Run service
resource "google_cloud_run_service_iam_member" "public_access" {
  location = google_cloud_run_v2_service.radiocast.location
  project  = google_cloud_run_v2_service.radiocast.project
  service  = google_cloud_run_v2_service.radiocast.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud Scheduler job
resource "google_cloud_scheduler_job" "daily_report" {
  name     = "radiocast-daily-report-${var.environment}"
  region   = var.region
  project  = var.project_id
  
  description = "Generate daily radio propagation report"
  schedule    = "0 0 * * *" # Daily at midnight UTC
  time_zone   = "UTC"
  
  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_v2_service.radiocast.uri}/generate"
    
    headers = {
      "Content-Type" = "application/json"
    }
    
    oidc_token {
      service_account_email = google_service_account.radiocast.email
    }
  }
  
  retry_config {
    retry_count          = var.environment == "production" ? 5 : 3
    max_retry_duration   = "300s"
    min_backoff_duration = "5s"
    max_backoff_duration = "60s"
    max_doublings        = 3
  }
  
  depends_on = [google_project_service.apis]
}

# Monitoring alert policy for production
resource "google_monitoring_alert_policy" "report_generation_failures" {
  count        = var.enable_monitoring ? 1 : 0
  display_name = "Radiocast Report Generation Failures"
  project      = var.project_id
  
  conditions {
    display_name = "Cloud Run service error rate"
    
    condition_threshold {
      filter         = "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"${var.service_name}\""
      duration       = "300s"
      comparison     = "COMPARISON_GREATER_THAN"
      threshold_value = 0.1
      
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }
  
  notification_channels = []
  
  alert_strategy {
    auto_close = "1800s"
  }
  
  depends_on = [google_project_service.apis]
}
