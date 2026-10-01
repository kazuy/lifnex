output "artifact_registry_app_repository_url" {
  description = "Artifact Registry Docker repository URL for the application."
  value       = "${google_artifact_registry_repository.app.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.app.repository_id}"
}

output "cloud_run_app_url" {
  description = "Cloud Run URL for the application."
  value       = google_cloud_run_v2_service.app.uri
}

output "github_actions_workload_identity_provider" {
  description = "Workload Identity Provider resource name used by GitHub Actions."
  value       = google_iam_workload_identity_pool_provider.github_actions.name
}

output "github_actions_pr_checks_service_account" {
  description = "Service account email used by pull request checks."
  value       = google_service_account.github_actions_pr_checks.email
}

output "github_actions_deploy_infra_service_account" {
  description = "Service account email used by approved Terraform applies."
  value       = google_service_account.github_actions_deploy_infra.email
}

output "github_actions_deploy_app_service_account" {
  description = "Service account email used by approved application deployments."
  value       = google_service_account.github_actions_deploy_app.email
}
