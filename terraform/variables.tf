variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "service_name" {
  description = "Name of the Cloud Run service"
  type        = string
}

variable "environment" {
  description = "Environment name (staging/production)"
  type        = string
}

variable "openai_api_key" {
  description = "OpenAI API key"
  type        = string
  sensitive   = true
}

variable "radiocast_api_key" {
  description = "Radiocast API key for /generate endpoint protection"
  type        = string
  sensitive   = true
}

variable "tfstate_bucket_name" {
  description = "Name of the GCS bucket for Terraform state"
  type        = string
}

variable "min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 10
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "latest"
}

variable "cpu_limit" {
  description = "CPU limit for Cloud Run service"
  type        = string
  default     = "2"
}

variable "memory_limit" {
  description = "Memory limit for Cloud Run service"
  type        = string
  default     = "2Gi"
}

variable "timeout" {
  description = "Request timeout for Cloud Run service"
  type        = string
  default     = "300s"
}


variable "enable_monitoring" {
  description = "Enable monitoring and alerting"
  type        = bool
  default     = false
}

variable "radiocast_bucket_name" {
  description = "Name of the GCS bucket for radiocast application (no public access)"
  type        = string
}

variable "radiocast_retention_days" {
  description = "Number of days to retain radiocast bucket objects in GCS"
  type        = number
  default     = 180
}

variable "github_actions_sa_email" {
  description = "Email of the GitHub Actions service account that needs to read secrets during deployment"
  type        = string
  default     = ""
}
