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
      image = var.image_uri

      env {
        name  = "TRANSPORT"
        value = "http"
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
