locals {
  github_actions_deploy_infra_roles = toset([
    "roles/artifactregistry.admin",
    "roles/iam.roleAdmin",
    "roles/iam.serviceAccountAdmin",
    "roles/iam.workloadIdentityPoolAdmin",
    "roles/resourcemanager.projectIamAdmin",
    "roles/run.admin",
    "roles/serviceusage.serviceUsageAdmin",
  ])
}

resource "google_service_account" "github_actions_deploy_infra" {
  project      = var.project_id
  account_id   = "github-actions-deploy-infra"
  display_name = "GitHub Actions infrastructure deployment"
  description  = "Terraform apply identity used after infrastructure approval."

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["iam.googleapis.com"],
  ]
}

resource "google_service_account_iam_member" "github_actions_deploy_infra_identity" {
  service_account_id = google_service_account.github_actions_deploy_infra.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${google_iam_workload_identity_pool.github_actions.name}/subject/repo:${var.github_repository}:environment:${var.github_infra_environment}"
}

resource "google_project_iam_member" "github_actions_deploy_infra" {
  for_each = local.github_actions_deploy_infra_roles

  project = var.project_id
  role    = each.value
  member  = google_service_account.github_actions_deploy_infra.member
}

resource "google_storage_bucket_iam_member" "github_actions_deploy_infra_state" {
  bucket = var.terraform_state_bucket
  role   = "roles/storage.objectAdmin"
  member = google_service_account.github_actions_deploy_infra.member
}

resource "google_service_account_iam_member" "github_actions_deploy_infra_runtime_user" {
  service_account_id = google_service_account.app.name
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.github_actions_deploy_infra.member
}

resource "google_service_account" "github_actions_deploy_app" {
  project      = var.project_id
  account_id   = "github-actions-deploy-app"
  display_name = "GitHub Actions application deployment"
  description  = "Cloud Run deployment identity used after application approval."

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["iam.googleapis.com"],
  ]
}

resource "google_service_account_iam_member" "github_actions_deploy_app_identity" {
  service_account_id = google_service_account.github_actions_deploy_app.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${google_iam_workload_identity_pool.github_actions.name}/subject/repo:${var.github_repository}:environment:${var.github_app_environment}"
}

resource "google_project_iam_member" "github_actions_deploy_app_cloud_run" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = google_service_account.github_actions_deploy_app.member
}

resource "google_artifact_registry_repository_iam_member" "github_actions_deploy_app_writer" {
  project    = google_artifact_registry_repository.app.project
  location   = google_artifact_registry_repository.app.location
  repository = google_artifact_registry_repository.app.repository_id
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.github_actions_deploy_app.member
}

resource "google_service_account_iam_member" "github_actions_deploy_app_runtime_user" {
  service_account_id = google_service_account.app.name
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.github_actions_deploy_app.member
}
