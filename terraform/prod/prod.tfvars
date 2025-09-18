project_id            = "dfh-prod-id"
service_name        = "radiocast-prod"
environment         = "production"
tfstate_bucket_name = "dfh-prod-tfstate"
radiocast_bucket_name = "dfh-prod-radiocast"
radiocast_retention_days = 180

# Resource limits for production
min_instances = 0
max_instances = 3
cpu_limit     = "2"
memory_limit  = "2Gi"
timeout       = "600s"


# Enable monitoring for production
enable_monitoring = true

# GitHub Actions service account for deployment
github_actions_sa_email = "radiocast-prod-deploy@dfh-prod-id.iam.gserviceaccount.com"
