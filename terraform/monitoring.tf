# Monitoring alert policy for production
resource "google_monitoring_alert_policy" "report_generation_failures" {
  count        = var.enable_monitoring ? 1 : 0
  display_name = "Radiocast Report Generation Failures"
  project      = var.project_id

  conditions {
    display_name = "Cloud Run service error rate"

    condition_threshold {
      filter          = "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"${var.service_name}\" AND metric.type=\"run.googleapis.com/request_count\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.1

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  combiner = "OR"

  notification_channels = []

  alert_strategy {
    auto_close = "1800s"
  }

  depends_on = [google_project_service.apis]
}
