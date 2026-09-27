output "artifact_registry_app_repository_url" {
  description = "Artifact Registry Docker repository URL for the application."
  value       = "${google_artifact_registry_repository.app.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.app.repository_id}"
}

output "cloud_run_app_url" {
  description = "Cloud Run URL for the application."
  value       = google_cloud_run_v2_service.app.uri
}
