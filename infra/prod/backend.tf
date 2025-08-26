terraform {
  backend "gcs" {
    bucket = "dfh-prod-tfstate"
    prefix = "terraform/state"
  }
}
