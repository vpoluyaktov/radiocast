output "service_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.radiocast.uri
}


output "service_account_email" {
  description = "Email of the service account"
  value       = google_service_account.radiocast.email
}

output "project_id" {
  description = "GCP project ID"
  value       = var.project_id
}

output "region" {
  description = "GCP region"
  value       = var.region
}
