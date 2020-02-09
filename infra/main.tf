provider "google" {
  credentials = file("account.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_storage_bucket" "functions_bucket" {
  name = "${var.project_name}_functions"
}

data "archive_file" "launch_zip" {
  type        = "zip"
  source_dir  = "${path.root}/fn/launch/"
  output_path = "${path.root}/launch.zip"
}

resource "google_storage_bucket_object" "launch_zip" {
  name   = "launch.zip"
  bucket = google_storage_bucket.functions_bucket.name
  source = "${path.root}/launch.zip"
}

resource "google_cloudfunctions_function" "launch_function" {
  name                  = "launch_function"
  available_memory_mb   = 256
  source_archive_bucket = google_storage_bucket.functions_bucket.name
  source_archive_object = google_storage_bucket_object.launch_zip.name
  entry_point           = "HelloHTTP"
  trigger_http          = true
  runtime               = "go111"
}
