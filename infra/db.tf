resource "google_compute_global_address" "sql_private_ip_address" {
  name          = "sql-private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = data.google_compute_network.default.self_link
}

resource "google_service_networking_connection" "sql_private_vpc_connection" {
  provider = google-beta

  network                 = data.google_compute_network.default.self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.sql_private_ip_address.name]
}


resource "google_sql_database_instance" "primary" {
  name             = "primary-instance"
  database_version = "POSTGRES_11"

  depends_on = [google_service_networking_connection.sql_private_vpc_connection]

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
    disk_autoresize   = true
    disk_type         = "PD_HDD"

    ip_configuration {
      ipv4_enabled    = false
      private_network = data.google_compute_network.default.self_link
    }
  }
}

resource "google_sql_database" "database" {
  name     = var.project_name
  instance = google_sql_database_instance.primary.name
}

resource "random_password" "sql_admin_password" {
  keepers = {
    name = google_sql_database_instance.primary.name
  }

  length  = 32
  special = true
}

resource "google_sql_user" "admin" {
  name     = "${var.project_name}_admin"
  instance = google_sql_database_instance.primary.name
  password = random_password.sql_admin_password.result
}

resource "google_secret_manager_secret" "sql_admin_password_secret" {
  provider  = google-beta
  secret_id = "sql_admin_password"

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "sql_admin_password_secret_version" {
  provider    = google-beta
  secret      = google_secret_manager_secret.sql_admin_password_secret.id
  secret_data = random_password.sql_admin_password.result
}
