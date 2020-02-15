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
