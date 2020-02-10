provider "google" {
  credentials = file("account.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_storage_bucket" "functions_bucket" {
  name = "${var.project_name}_functions"
}

data "archive_file" "function_zip" {
  for_each = toset(var.functions)

  type        = "zip"
  source_dir  = "${var.functions_path}/${each.key}/"
  output_path = "${path.root}/${each.key}.zip"
}

resource "google_storage_bucket_object" "function_zip" {
  for_each = toset(var.functions)

  name   = "${each.key}.zip"
  bucket = google_storage_bucket.functions_bucket.name
  source = "${path.root}/${each.key}.zip"
}

resource "google_cloudfunctions_function" "launch_function" {
  name                  = "launch_function"
  available_memory_mb   = 256
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.function_zip["launch"].name
  entry_point           = "HelloHTTP"
  trigger_http          = true
  runtime               = "go111"
}
