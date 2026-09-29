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
