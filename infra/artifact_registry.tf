resource "google_artifact_registry_repository" "app" {
  location      = var.region
  repository_id = "${var.project_id}-app"
  format        = "DOCKER"

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["artifactregistry.googleapis.com"],
  ]
}
