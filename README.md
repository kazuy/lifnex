# lifnex

## About

**lifnex** (pronounced "life-nex") is a coined name combining **life** and **nexus**.

> Connect everyday life services, local data, and personal tools to AI assistants.

## Architecture

```mermaid
flowchart LR
    User["User"]
    Client["ChatGPT / MCP Client"]
    Auth["External OAuth 2.1<br/>Authorization Server"]
    Events["Kawasaki City<br/>Event Data"]

    subgraph GitHub["GitHub"]
        Actions["GitHub Actions<br/>PR Checks / Deploy"]
    end

    subgraph GoogleCloud["Google Cloud"]
        WIF["Workload Identity<br/>Federation"]
        State["GCS Terraform State"]
        Registry["Artifact Registry"]
        Run["Cloud Run<br/>lifnex MCP Server"]
    end

    User --> Client
    Client -->|"Sign-in / consent"| Auth
    Auth -->|"Access token"| Client
    Client -->|"HTTPS / MCP"| Run
    Run -->|"Verify token with JWKS"| Auth
    Run -->|"Search events"| Events

    Actions -->|"OIDC"| WIF
    WIF -. "Short-lived credentials" .-> Actions
    Actions -->|"Terraform plan / apply"| GoogleCloud
    Actions -->|"Push image"| Registry
    Actions -->|"Deploy revision"| Run
    Registry -->|"Container image"| Run
```

## Infrastructure

See [Infrastructure](infra/README.md) to provision the initial Google Cloud
environment or to plan and apply later Terraform changes.
