locals {
  github_repository_parts = split("/", var.github_repository)
  github_repository_subject = join("/", [
    "repo:${local.github_repository_parts[0]}@${var.github_repository_owner_id}",
    "${local.github_repository_parts[1]}@${var.github_repository_id}",
  ])
}

resource "google_iam_workload_identity_pool" "github_actions" {
  project                   = var.project_id
  workload_identity_pool_id = "github-actions"
  display_name              = "GitHub Actions"
  description               = "Federated identities for ${var.github_repository} GitHub Actions workflows."

  deletion_policy = "PREVENT"

  depends_on = [
    google_project_service.api["iam.googleapis.com"],
  ]
}

resource "google_iam_workload_identity_pool_provider" "github_actions" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github_actions.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub"

  deletion_policy = "PREVENT"

  attribute_mapping = {
    "google.subject"          = "assertion.sub"
    "attribute.repository_id" = "assertion.repository_id"
  }
  attribute_condition = "assertion.repository_id == '${var.github_repository_id}'"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}
