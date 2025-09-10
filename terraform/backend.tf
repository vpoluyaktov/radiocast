terraform {
  backend "gcs" {
    bucket = "dfh-prod-tfstate"
    prefix = "radiocast/state"
  }
}
