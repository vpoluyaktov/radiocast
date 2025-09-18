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

# GitHub Actions service account permissions for secret access during deployment
resource "google_project_iam_member" "github_actions_secrets" {
  count   = var.github_actions_sa_email != "" ? 1 : 0
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${var.github_actions_sa_email}"
}
