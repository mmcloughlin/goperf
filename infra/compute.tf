resource "google_compute_project_default_network_tier" "default" {
  network_tier = var.network_tier
}

data "google_compute_image" "ubuntu" {
  family  = "ubuntu-1804-lts"
  project = "ubuntu-os-cloud"
}

resource "google_storage_bucket" "artifacts_bucket" {
  name = "${var.project_name}_artifacts"
}

locals {
  dist_archive_path   = "${path.root}/${var.dist_path}"
  dist_archive_sha256 = filesha256(local.dist_archive_path)
}


resource "google_storage_bucket_object" "dist_archive" {
  name   = "${var.project_name}/${local.dist_archive_sha256}.tar.gz"
  bucket = google_storage_bucket.artifacts_bucket.name
  source = local.dist_archive_path
}

resource "google_compute_instance" "worker" {
  name                      = "worker"
  machine_type              = var.worker_machine_type
  allow_stopping_for_update = true

  metadata_startup_script = templatefile("${path.root}/init.sh", {
    deploy_dir          = "/opt/${var.project_name}",
    dist_archive_gs_uri = "${google_storage_bucket.artifacts_bucket.url}/${google_storage_bucket_object.dist_archive.name}",
  })

  service_account {
    email  = data.google_service_account.bot.email
    scopes = ["pubsub", "storage-rw"]
  }

  boot_disk {
    initialize_params {
      image = data.google_compute_image.ubuntu.self_link
    }
  }

  network_interface {
    network = "default"
    access_config {
      network_tier = var.network_tier
    }
  }
}
