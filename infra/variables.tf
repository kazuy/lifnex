variable "project_id" {
  description = "Google Cloud project ID."
  type        = string
}

variable "region" {
  description = "Google Cloud region for regional resources."
  type        = string
}

variable "image_uri" {
  description = "Initial container image URI for the Cloud Run service."
  type        = string
}
