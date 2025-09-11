project_id            = "dfh-stage-id"
service_name        = "radiocast-stage"
environment         = "staging"
tfstate_bucket_name = "dfh-stage-tfstate"
radiocast_bucket_name = "dfh-stage-radiocast"
radiocast_retention_days = 180

# Resource limits for staging
min_instances = 0
max_instances = 3
cpu_limit     = "2"
memory_limit  = "2Gi"
timeout       = "300s"


# Monitoring disabled for staging
enable_monitoring = false
