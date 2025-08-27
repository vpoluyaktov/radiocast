project_id            = "dfh-stage-id"
service_name        = "radiocast-stage"
environment         = "staging"
reports_bucket_name = "dfh-stage-reports"
tfstate_bucket_name = "dfh-stage-tfstate"

# Resource limits for staging
min_instances = 0
max_instances = 10
cpu_limit     = "2"
memory_limit  = "2Gi"
timeout       = "300s"

# Retention
reports_retention_days = 90

# Monitoring disabled for staging
enable_monitoring = false
