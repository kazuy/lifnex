resource "google_cloud_run_v2_service" "app" {
  name     = "${var.project_id}-app"
  location = var.region

  ingress = "INGRESS_TRAFFIC_ALL"

  invoker_iam_disabled = true

  deletion_protection = true

  template {
    service_account = google_service_account.app.email

    scaling {
      max_instance_count = 1
    }

    containers {
      image = var.bootstrap_image_uri

      env {
        name  = "TRANSPORT"
        value = "http"
      }

      env {
        name  = "OAUTH_ISSUER_URL"
        value = var.oauth_issuer_url
      }

      env {
        name  = "OAUTH_JWKS_URL"
        value = var.oauth_jwks_url
      }

      env {
        name  = "OAUTH_RESOURCE_URL"
        value = var.oauth_resource_url
      }

      ports {
        container_port = 8080
      }

      startup_probe {
        http_get {
          path = "/healthz"
        }
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
    ]
  }

  depends_on = [
    google_artifact_registry_repository.app,
    google_project_service.api["run.googleapis.com"],
  ]
}
