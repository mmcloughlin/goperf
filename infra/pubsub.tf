resource "google_pubsub_topic" "jobs" {
  name = var.jobs_topic
}

resource "google_pubsub_subscription" "worker_jobs_subscription" {
  name  = "worker_${google_pubsub_topic.jobs.name}"
  topic = google_pubsub_topic.jobs.name

  message_retention_duration = "3600s"
  retain_acked_messages      = true
  ack_deadline_seconds       = 30
}
