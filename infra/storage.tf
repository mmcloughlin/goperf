resource "google_storage_bucket" "results_bucket" {
  name = "${var.project_name}_results"
}
