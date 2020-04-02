output "db_connection_name" {
  value = google_sql_database_instance.primary.connection_name
}

output "db_name" {
  value = google_sql_database.database.name
}

output "db_user" {
  value = google_sql_user.admin.name
}

output "db_password" {
  value     = google_sql_user.admin.password
  sensitive = true
}
