variable "project_id" {
  description = "Google Cloud project ID."
  type        = string
}

variable "region" {
  description = "Google Cloud region for regional resources."
  type        = string
}

variable "bootstrap_image_uri" {
  description = "Container image URI used only when initially creating the Cloud Run service."
  type        = string
}

variable "terraform_state_bucket" {
  description = "Name of the pre-created GCS bucket that stores Terraform state."
  type        = string
}

variable "github_repository" {
  description = "GitHub repository allowed to authenticate through Workload Identity Federation."
  type        = string
  default     = "kazuy/lifnex"

  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repository))
    error_message = "github_repository must use the owner/repository format."
  }
}

variable "github_infra_environment" {
  description = "Protected GitHub environment allowed to use the Terraform apply identity."
  type        = string
  default     = "production-infra"

  validation {
    condition     = length(trimspace(var.github_infra_environment)) > 0
    error_message = "github_infra_environment must not be empty."
  }
}

variable "github_app_environment" {
  description = "Protected GitHub environment allowed to use the Cloud Run deployment identity."
  type        = string
  default     = "production-app"

  validation {
    condition     = length(trimspace(var.github_app_environment)) > 0
    error_message = "github_app_environment must not be empty."
  }
}

variable "oauth_issuer_url" {
  description = "OAuth authorization server issuer URL expected in access tokens."
  type        = string

  validation {
    condition     = can(regex("^https://[^/?#]+", var.oauth_issuer_url))
    error_message = "oauth_issuer_url must be an absolute HTTPS URL."
  }
}

variable "oauth_jwks_url" {
  description = "OAuth authorization server JWKS URL used to verify access tokens."
  type        = string

  validation {
    condition     = can(regex("^https://[^/?#]+", var.oauth_jwks_url))
    error_message = "oauth_jwks_url must be an absolute HTTPS URL."
  }
}

variable "oauth_resource_url" {
  description = "Canonical MCP resource URL expected in the access token audience."
  type        = string

  validation {
    condition     = can(regex("^https://[^/?#]+", var.oauth_resource_url)) && endswith(var.oauth_resource_url, "/mcp")
    error_message = "oauth_resource_url must be an absolute HTTPS URL ending in /mcp."
  }
}
