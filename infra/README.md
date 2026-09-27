# Infrastructure

## Infrastructure diagram

```mermaid
flowchart LR
    Client["ChatGPT / MCP Client"]
    Developer["Developer"]

    subgraph Existing["Pre-created resources"]
        Project["Google Cloud Project"]
        State["GCS Terraform State"]
    end

    subgraph Terraform["Terraform-managed resources"]
        APIs["Google Cloud APIs"]
        AR["Artifact Registry"]
        SA["Cloud Run Service Account"]
        Run["Cloud Run Service"]
        IAM["IAM Binding<br/>allUsers → roles/run.invoker"]

        APIs --> AR
        APIs --> SA
        APIs --> Run
        AR -->|"Container image"| Run
        SA -->|"Runtime identity"| Run
        IAM -.->|"Attached to"| Run
    end

    Developer -->|"Terraform state"| State
    Developer -->|"Terraform apply"| APIs
    Developer -->|"Build and push image"| AR
    Project --- Terraform
    Client -->|"HTTPS / MCP"| Run
```

The Google Cloud project and Terraform state bucket are created separately.
Remote MCP clients connect directly to the standard Cloud Run HTTPS endpoint.

## Initial setup

Run the following commands from this directory.

Authenticate with Google Cloud Application Default Credentials, then initialize
Terraform with the separately created state bucket:

```sh
gcloud auth application-default login

cp terraform.tfvars.example terraform.tfvars
terraform init -backend-config="bucket=<TFSTATE_BUCKET>"
```

Replace the example values in `terraform.tfvars` for the target environment.

## Apply changes

Review the plan, then apply the changes:

```sh
terraform plan
terraform apply
```
