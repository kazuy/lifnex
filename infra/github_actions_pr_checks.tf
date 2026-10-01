locals {
  github_actions_pr_checks_roles = toset([
    "roles/iam.roleViewer",
    "roles/serviceusage.serviceUsageConsumer",
    "roles/viewer",
  ])
}

resource "google_service_account" "github_actions_pr_checks" {
  project      = var.project_id
  account_id   = "github-actions-pr-checks"
  display_name = "GitHub Actions PR checks"
  description  = "Read-only identity used by pull request checks."

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["iam.googleapis.com"],
  ]
}

resource "google_service_account_iam_member" "github_actions_pr_checks_identity" {
  service_account_id = google_service_account.github_actions_pr_checks.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_actions.name}/attribute.repository/${var.github_repository}"
}

resource "google_project_iam_custom_role" "github_actions_pr_checks_state_reader" {
  project     = var.project_id
  role_id     = "githubActionsTerraformStateReader"
  title       = "GitHub Actions Terraform state reader"
  description = "Read-only access to Terraform state and its bucket IAM policy."

  deletion_policy = "PREVENT"

  permissions = [
    "storage.buckets.get",
    "storage.buckets.getIamPolicy",
    "storage.objects.get",
    "storage.objects.list",
  ]
}

resource "google_project_iam_member" "github_actions_pr_checks" {
  for_each = local.github_actions_pr_checks_roles

  project = var.project_id
  role    = each.value
  member  = google_service_account.github_actions_pr_checks.member
}

resource "google_storage_bucket_iam_member" "github_actions_pr_checks_state" {
  bucket = var.terraform_state_bucket
  role   = google_project_iam_custom_role.github_actions_pr_checks_state_reader.name
  member = google_service_account.github_actions_pr_checks.member
}
