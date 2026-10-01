# Infrastructure

Terraform provisions the Google Cloud APIs, Artifact Registry repository,
service accounts, Cloud Run service, and GitHub Actions identities used by
lifnex. The Google Cloud project and GCS state bucket must already exist.

The OAuth authorization server is external and is not managed here. Before
deploying, configure it to issue JWT access tokens whose issuer and audience
exactly match `oauth_issuer_url` and `oauth_resource_url`.

## Prerequisites

- Google Cloud CLI and Docker
- A Google Cloud project and GCS Terraform state bucket
- An external OAuth 2.1 authorization server and its JWKS URL
- Permission to manage the resources and IAM bindings in this configuration

Run all commands below from the repository root. Install the Terraform version
declared in `mise.toml`, then authenticate locally:

```sh
mise trust
mise install
gcloud auth application-default login
```

## Initial provisioning

Create the local variables file and set every value. Use the same state bucket
name for `terraform_state_bucket` and the backend configuration.

```sh
cp infra/terraform.tfvars.example infra/terraform.tfvars
mise exec -- terraform -chdir=infra init \
  -backend-config="bucket=<TFSTATE_BUCKET>"
```

The first container image cannot be pushed until Artifact Registry exists.
Create only the repository and its dependencies for this bootstrap step:

```sh
mise exec -- terraform -chdir=infra plan \
  -target=google_artifact_registry_repository.app \
  -out=bootstrap-registry.tfplan
mise exec -- terraform -chdir=infra apply bootstrap-registry.tfplan
```

Build and push the initial image:

```sh
APP_REPOSITORY_URL="$(
  mise exec -- terraform -chdir=infra output -raw \
    artifact_registry_app_repository_url
)"
REGISTRY_HOST="${APP_REPOSITORY_URL%%/*}"
IMAGE_URI="${APP_REPOSITORY_URL}/lifnex:$(git rev-parse --short HEAD)"

gcloud auth configure-docker "${REGISTRY_HOST}"
make build-image IMAGE_URI="${IMAGE_URI}"
docker push "${IMAGE_URI}"
```

Set `bootstrap_image_uri` in `infra/terraform.tfvars` to the pushed image URI.
Terraform uses this image only when initially creating the Cloud Run service.
Review and apply the complete configuration; routine changes must not use
`-target`.

```sh
mise exec -- terraform -chdir=infra plan -out=bootstrap.tfplan
mise exec -- terraform -chdir=infra apply bootstrap.tfplan
```

## Configure GitHub Actions

The initial apply creates Workload Identity Federation and separate service
accounts for read-only PR checks, Terraform applies, and application
deployments. GitHub Actions uses short-lived OIDC credentials, so no service
account key is stored in GitHub.

Read the generated identifiers:

```sh
mise exec -- terraform -chdir=infra output -raw \
  github_actions_workload_identity_provider
mise exec -- terraform -chdir=infra output -raw \
  github_actions_pr_checks_service_account
mise exec -- terraform -chdir=infra output -raw \
  github_actions_deploy_infra_service_account
mise exec -- terraform -chdir=infra output -raw \
  github_actions_deploy_app_service_account
```

Configure these repository secrets. They are identifiers and configuration
values rather than credentials, but keeping them as secrets prevents accidental
disclosure in workflow logs:

| Variable | Value |
| --- | --- |
| `GCP_PROJECT_ID` | `project_id` from `terraform.tfvars` |
| `GCP_REGION` | `region` from `terraform.tfvars` |
| `GCP_BOOTSTRAP_IMAGE_URI` | `bootstrap_image_uri` from `terraform.tfvars` |
| `TF_STATE_BUCKET` | `terraform_state_bucket` from `terraform.tfvars` |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `github_actions_workload_identity_provider` output |
| `GCP_PR_CHECKS_SERVICE_ACCOUNT` | `github_actions_pr_checks_service_account` output |
| `OAUTH_ISSUER_URL` | `oauth_issuer_url` from `terraform.tfvars` |
| `OAUTH_JWKS_URL` | `oauth_jwks_url` from `terraform.tfvars` |
| `OAUTH_RESOURCE_URL` | `oauth_resource_url` from `terraform.tfvars` |

Create two protected GitHub environments with required reviewers, deployment
branch policies, and protection-rule bypass disabled:

| Environment | Environment secret | Value |
| --- | --- | --- |
| `production-infra` | `GCP_INFRA_SERVICE_ACCOUNT` | `github_actions_deploy_infra_service_account` output |
| `production-app` | `GCP_APP_SERVICE_ACCOUNT` | `github_actions_deploy_app_service_account` output |

Allow only `main` for `production-infra`. For `production-app`, allow `main`
and any branch or tag that may be selected for a manual application deployment.
Set `github_infra_environment` or `github_app_environment` before applying
Terraform if different environment names are required.

Pushes to `main` inspect the changed paths. An `infra/` change creates a
read-only Terraform plan and publishes its change counts to the job summary. A
zero-change plan finishes without requesting approval; otherwise,
`production-infra` approval gates Terraform apply. An `app/` or `Makefile`
change independently requests `production-app` approval before publishing the
image and updating Cloud Run. Neither deployment waits for the other.

A manual workflow run deploys only the application from the branch or tag
selected in GitHub's standard workflow form. Infrastructure is intentionally
limited to changes pushed to `main`.

Application deployments build a commit-SHA-tagged image URI in the workflow.
Terraform ignores subsequent changes to the Cloud Run container image, so the
fixed bootstrap image is not used to roll back application deployments.

## Apply later changes

After cloning an already provisioned environment, copy and complete
`infra/terraform.tfvars`, authenticate, and initialize Terraform as described
above. Then save the reviewed plan and apply that exact plan:

```sh
mise exec -- terraform -chdir=infra plan -out=terraform.tfplan
mise exec -- terraform -chdir=infra apply terraform.tfplan
```
