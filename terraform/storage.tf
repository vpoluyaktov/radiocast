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
        with_state         = "ARCHIVED"
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
