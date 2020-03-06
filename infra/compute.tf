resource "google_compute_project_default_network_tier" "default" {
  network_tier = var.network_tier
}

data "google_compute_image" "ubuntu" {
  family  = "ubuntu-1804-lts"
  project = "ubuntu-os-cloud"
}

resource "google_storage_bucket" "results_bucket" {
  name = "${var.project_name}_results"
}

resource "google_storage_bucket" "artifacts_bucket" {
  name = "${var.project_name}_artifacts"
}

locals {
  dist_archive_path   = "dist.tar.gz"
  dist_archive_sha256 = filesha256(local.dist_archive_path)
}

resource "google_storage_bucket_object" "dist_archive" {
  name   = "${var.project_name}/${local.dist_archive_sha256}.tar.gz"
  bucket = google_storage_bucket.artifacts_bucket.name
  source = local.dist_archive_path
}

resource "google_compute_instance_template" "worker" {
  name_prefix      = "worker-"
  machine_type     = var.worker_machine_type
  min_cpu_platform = var.worker_min_cpu_platform

  metadata_startup_script = templatefile("${path.root}/init.sh", {
    project_name        = var.project_name,
    deploy_dir          = "/opt/${var.project_name}",
    log_dir             = "/var/log/${var.project_name}",
    dist_archive_gs_uri = "${google_storage_bucket.artifacts_bucket.url}/${google_storage_bucket_object.dist_archive.name}",
  })

  scheduling {
    preemptible       = true
    automatic_restart = false
  }

  service_account {
    email  = data.google_service_account.bot.email
    scopes = ["pubsub", "storage-rw"]
  }

  disk {
    source_image = data.google_compute_image.ubuntu.self_link
  }

  network_interface {
    network = "default"
    access_config {
      network_tier = var.network_tier
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_target_pool" "workers" {
  name = "workers"
}

resource "google_compute_instance_group_manager" "workers" {
  name = "workers"

  version {
    instance_template = google_compute_instance_template.worker.self_link
    name              = "primary"
  }

  target_pools       = [google_compute_target_pool.workers.self_link]
  base_instance_name = "worker"
}

resource "google_compute_autoscaler" "workers" {
  provider = google-beta

  name   = "workers"
  target = google_compute_instance_group_manager.workers.self_link

  autoscaling_policy {
    max_replicas    = 5
    min_replicas    = 0
    cooldown_period = 60

    metric {
      name                       = "pubsub.googleapis.com/subscription/num_undelivered_messages"
      filter                     = "resource.type = \"pubsub_subscription\" AND resource.label.subscription_id = \"${google_pubsub_subscription.worker_jobs_subscription.name}\""
      single_instance_assignment = 65535
    }
  }
}
