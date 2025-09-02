project_id            = "dfh-prod-id"
service_name        = "radiocast-prod"
environment         = "production"
reports_bucket_name = "dfh-prod-reports"
tfstate_bucket_name = "dfh-prod-tfstate"

# Resource limits for production
min_instances = 0
max_instances = 3
cpu_limit     = "2"
memory_limit  = "2Gi"
timeout       = "600s"

# Retention - keep production reports longer
reports_retention_days = 365

# Enable monitoring for production
enable_monitoring = true
