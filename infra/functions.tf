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
    CB_PROJECT_ID               = var.project_id
    CB_JOBS_TOPIC               = google_pubsub_topic.jobs.id
    CB_SQL_CONNECTION_NAME      = google_sql_database_instance.primary.connection_name
    CB_SQL_DATABASE             = google_sql_database.database.name
    CB_SQL_USER                 = google_sql_user.admin.name
    CB_SQL_PASSWORD_SECRET_NAME = google_secret_manager_secret_version.sql_admin_password_secret_version.name
  }
}

resource "google_cloudfunctions_function" "http_function" {
  for_each = toset([for f in var.functions : f.name if f.trigger_type == "http"])

  name                  = each.key
  available_memory_mb   = 128
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip[each.key].name
  entry_point           = "Handle"
  trigger_http          = true
  runtime               = var.functions_runtime
  environment_variables = local.environment_variables
}

resource "google_cloudfunctions_function" "enqueue_function" {
  name                  = "enqueue"
  available_memory_mb   = 128
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip["enqueue"].name
  entry_point           = "Handle"
  runtime               = var.functions_runtime
  environment_variables = local.environment_variables

  event_trigger {
    event_type = "providers/cloud.firestore/eventTypes/document.create"
    resource   = "${var.commits_collection}/{sha}"
  }
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

  event_trigger {
    event_type = "google.storage.object.finalize"
    resource   = google_storage_bucket.results_bucket.name
  }
}
