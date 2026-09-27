output "artifact_registry_app_repository_url" {
  description = "Artifact Registry Docker repository URL for the application."
  value       = "${google_artifact_registry_repository.app.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.app.repository_id}"
}
