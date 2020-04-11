provider "google" {
  credentials = file("account.json")
  project     = var.project_id
  region      = var.region
  zone        = var.zone
}

provider "google-beta" {
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

resource "google_compute_project_default_network_tier" "default" {
  network_tier = var.network_tier
}
