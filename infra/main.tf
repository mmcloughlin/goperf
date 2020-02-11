provider "google" {
  credentials = file("account.json")
  project     = var.project_id
  region      = var.region
  zone        = var.zone
}

resource "google_storage_bucket" "functions_bucket" {
  name = "${var.project_name}_functions"
}

data "archive_file" "function_zip" {
  for_each = toset(var.functions)

  type        = "zip"
  source_dir  = "${path.root}/fn/${each.key}/"
  output_path = "${path.root}/tmp/${each.key}.zip"
}

resource "google_storage_bucket_object" "function_zip" {
  for_each = toset(var.functions)

  name   = "${each.key}/${data.archive_file.function_zip[each.key].output_sha}.zip"
  bucket = google_storage_bucket.functions_bucket.name
  source = "${path.root}/tmp/${each.key}.zip"
}

resource "google_cloudfunctions_function" "http_function" {
  for_each = toset(var.functions)

  name                  = each.key
  available_memory_mb   = 128
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip[each.key].name
  entry_point           = "Handle"
  trigger_http          = true
  runtime               = var.functions_runtime

  environment_variables = {
    CB_PROJECT_ID = var.project_id
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
