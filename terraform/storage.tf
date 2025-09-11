
# GCS bucket for radiocast application (no public access)
resource "google_storage_bucket" "radiocast" {
  name     = var.radiocast_bucket_name
  location = var.region
  project  = var.project_id

  uniform_bucket_level_access = true

  versioning {
    enabled = var.environment == "production"
  }

  lifecycle_rule {
    condition {
      age = var.radiocast_retention_days
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

# Grant Cloud Run service account access to radiocast bucket
resource "google_storage_bucket_iam_member" "radiocast_service_access" {
  bucket = google_storage_bucket.radiocast.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.radiocast.email}"
}
