provider "google" {
  credentials = file("account.json")
  project     = var.project_id
  region      = var.region
  zone        = var.zone
}

terraform {
  backend "gcs" {
    bucket = "contbench_terraform_state"
  }
}

data "google_service_account" "bot" {
  account_id = var.service_account_id
}
