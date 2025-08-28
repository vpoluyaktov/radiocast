terraform {
  backend "gcs" {
    bucket = "dfh-stage-tfstate"
    prefix = "radiocast/state"
  }
}
