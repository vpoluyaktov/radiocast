project_id          = "dfh-prod"
service_name        = "radiocast-prod"
environment         = "production"
reports_bucket_name = "dfh-prod-reports"
tfstate_bucket_name = "dfh-prod-tfstate"

# Resource limits for production
min_instances = 1
max_instances = 20
cpu_limit     = "4"
memory_limit  = "4Gi"
timeout       = "600s"

# Retention - keep production reports longer
reports_retention_days = 365

# Enable monitoring for production
enable_monitoring = true
