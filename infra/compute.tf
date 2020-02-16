resource "google_compute_instance" "adhoc" {
  name         = "adhoc"
  machine_type = "f1-micro"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-1804-bionic-v20200129a"
    }
  }

  network_interface {
    network = "default"
  }
}

resource "google_storage_bucket" "artifacts_bucket" {
  name = "${var.project_name}_artifacts"
}

resource "google_storage_bucket_object" "dist_archive" {
  name   = "${var.project_name}.tar.gz"
  bucket = google_storage_bucket.artifacts_bucket.name
  source = "${path.root}/${var.dist_path}"
}
