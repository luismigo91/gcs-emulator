provider "google" {
  project      = "test-project"
  access_token = "ya29.emulator.mock.access.token"
  region       = "us-central1"
}

provider "google-beta" {
  project      = "test-project"
  access_token = "ya29.emulator.mock.access.token"
  region       = "us-central1"
}

resource "google_storage_bucket" "data" {
  name          = "my-data-bucket"
  location      = "US"
  storage_class = "STANDARD"

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = 30
    }
  }
}

resource "google_storage_bucket_object" "config" {
  name   = "config/settings.json"
  bucket = google_storage_bucket.data.name
  content = jsonencode({
    environment = "development"
    debug       = true
  })
}

resource "google_pubsub_topic" "events" {
  name = "app-events"

  message_retention_duration = "86600s"
}

resource "google_pubsub_subscription" "events_processor" {
  name  = "app-events-processor"
  topic = google_pubsub_topic.events.name

  ack_deadline_seconds = 20

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.events_dlq.id
    max_delivery_attempts = 5
  }
}

resource "google_pubsub_topic" "events_dlq" {
  name = "app-events-dlq"
}

resource "google_secret_manager_secret" "api_key" {
  secret_id = "api-key"
  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "api_key_v1" {
  secret      = google_secret_manager_secret.api_key.id
  secret_data = "sk-abcdef123456"
}

resource "google_service_account" "worker" {
  account_id   = "worker-sa"
  display_name = "Worker Service Account"
}
