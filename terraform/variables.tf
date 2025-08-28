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

variable "reports_bucket_name" {
  description = "Name of the GCS bucket for reports"
  type        = string
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

variable "reports_retention_days" {
  description = "Number of days to retain reports in GCS"
  type        = number
  default     = 90
}

variable "enable_monitoring" {
  description = "Enable monitoring and alerting"
  type        = bool
  default     = false
}
