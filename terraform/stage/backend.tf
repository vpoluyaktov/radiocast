terraform {
  backend "gcs" {
    bucket = "dfh-stage-tfstate"
    prefix = "terraform/state"
  }
}
