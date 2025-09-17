# Cloud Scheduler job
resource "google_cloud_scheduler_job" "daily_report" {
  name    = "radiocast-daily-report-${var.environment}"
  region  = var.region
  project = var.project_id

  description = "Generate daily radio propagation report"
  schedule    = "0 0 * * *" # Daily at midnight UTC
  time_zone   = "UTC"

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_v2_service.radiocast.uri}/generate"

    headers = {
      "Content-Type"  = "application/json"
      "Authorization" = "Bearer ${var.radiocast_api_key}"
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
