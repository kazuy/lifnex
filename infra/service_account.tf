resource "google_service_account" "app" {
  project      = var.project_id
  account_id   = substr("${var.project_id}-run", 0, 30)
  display_name = "Cloud Run runtime"

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["iam.googleapis.com"],
  ]
}
