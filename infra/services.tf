locals {
  required_apis = toset([
    "artifactregistry.googleapis.com",
    "iam.googleapis.com",
    "run.googleapis.com",
  ])
}

resource "google_project_service" "api" {
  for_each = local.required_apis

  project = var.project_id
  service = each.value

  disable_on_destroy = false
}
