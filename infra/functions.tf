resource "google_vpc_access_connector" "connector" {
  name          = "connector"
  region        = var.region
  ip_cidr_range = "10.8.0.0/28"
  network       = data.google_compute_network.default.name
}

resource "google_storage_bucket" "functions_bucket" {
  name = "${var.project_name}_functions"
}

data "archive_file" "function_zip" {
  for_each = toset(var.functions[*].name)

  type        = "zip"
  source_dir  = "fn/${each.key}/"
  output_path = "tmp/${each.key}.zip"
}

resource "google_storage_bucket_object" "function_zip" {
  for_each = toset(var.functions[*].name)

  name   = "${each.key}/${data.archive_file.function_zip[each.key].output_sha}.zip"
  bucket = google_storage_bucket.functions_bucket.name
  source = data.archive_file.function_zip[each.key].output_path
}

locals {
  environment_variables = {
    CB_SQL_IP_ADDRESS           = google_sql_database_instance.primary.private_ip_address
    CB_SQL_DATABASE             = google_sql_database.database.name
    CB_SQL_USER                 = google_sql_user.admin.name
    CB_SQL_PASSWORD_SECRET_NAME = google_secret_manager_secret_version.sql_admin_password_secret_version.name
    CB_RESULTS_BUCKET           = google_storage_bucket.results_bucket.name
  }
}

resource "google_cloudfunctions_function" "http_function" {
  for_each = toset([for f in var.functions : f.name if f.trigger_type == "http"])

  name                  = each.key
  available_memory_mb   = 256
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip[each.key].name
  entry_point           = "Handle"
  trigger_http          = true
  runtime               = var.functions_runtime
  environment_variables = local.environment_variables
  vpc_connector         = google_vpc_access_connector.connector.id
}

resource "google_cloudfunctions_function_iam_member" "invoker" {
  for_each = toset([for f in var.functions : f.name if f.trigger_type == "http"])

  cloud_function = google_cloudfunctions_function.http_function[each.key].name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

resource "google_cloud_scheduler_job" "watch_schedule" {
  name             = "schedule_watch"
  schedule         = "13 * * * *"
  time_zone        = "Etc/UTC"
  attempt_deadline = "120s"

  http_target {
    http_method = "GET"
    uri         = google_cloudfunctions_function.http_function["watch"].https_trigger_url
  }
}

resource "google_cloud_scheduler_job" "staletimeout_schedule" {
  name             = "schedule_staletimeout"
  schedule         = "*/5 * * * *"
  time_zone        = "Etc/UTC"
  attempt_deadline = "120s"

  http_target {
    http_method = "GET"
    uri         = google_cloudfunctions_function.http_function["staletimeout"].https_trigger_url
  }
}

resource "google_cloudfunctions_function" "result_function" {
  for_each = toset([for f in var.functions : f.name if f.trigger_type == "result"])

  name                  = each.key
  available_memory_mb   = 256
  timeout               = 480
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip[each.key].name
  entry_point           = "Handle"
  runtime               = var.functions_runtime
  environment_variables = local.environment_variables
  vpc_connector         = google_vpc_access_connector.connector.id

  event_trigger {
    event_type = "google.storage.object.finalize"
    resource   = google_storage_bucket.results_bucket.name
  }
}
