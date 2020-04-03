resource "google_sql_database_instance" "primary" {
  name             = "primary-instance"
  database_version = "POSTGRES_11"

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
    disk_autoresize   = true
    disk_type         = "PD_HDD"
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
