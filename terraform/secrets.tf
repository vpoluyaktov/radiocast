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

# Secret for Radiocast API key
resource "google_secret_manager_secret" "radiocast_api_key" {
  secret_id = "radiocast-api-key"
  project   = var.project_id

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "radiocast_api_key" {
  secret      = google_secret_manager_secret.radiocast_api_key.id
  secret_data = var.radiocast_api_key
}
