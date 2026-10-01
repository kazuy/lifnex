# lifnex

**lifnex** (pronounced "life-nex") is a personal remote MCP server that
connects AI assistants with local and everyday-life information.

The current implementation provides authenticated access to Kawasaki City
event data and runs on Google Cloud Run.

## MCP Tools

| Tool | Description |
| --- | --- |
| `hello` | Return a hello-world greeting. |
| `search_events` | Search Kawasaki City events by date range, title keyword, and area. |

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

lifnex acts as an OAuth Resource Server and validates JWT access tokens against
the external authorization server's JWKS.

## Endpoints

| Path | Authentication | Purpose |
| --- | --- | --- |
| `/mcp` | Bearer token | Streamable HTTP MCP endpoint |
| `/.well-known/oauth-protected-resource` | None | OAuth Protected Resource Metadata |
| `/healthz` | None | Internal Cloud Run startup probe |

## CI/CD

- Pull requests run checks only for the affected application or infrastructure.
- Terraform plans use a read-only identity.
- Infrastructure and application deployments use separate identities and approval flows.
- GitHub Actions authenticates to Google Cloud with OIDC and Workload Identity Federation.

## Local Development

Install the tool versions declared in `mise.toml`:

```sh
mise trust
mise install
```

Run the application checks:

```sh
mise exec -- make go-format-check
mise exec -- make go-vet
mise exec -- make go-test
mise exec -- make go-build
```

## Infrastructure

See [Infrastructure](infra/README.md) for Google Cloud provisioning, GitHub
Actions configuration, and ongoing Terraform operations.

## Tech Stack

| Area | Technologies |
| --- | --- |
| Application | Go, MCP Go SDK |
| Authentication | OAuth 2.1, JWT, JWKS |
| Infrastructure | Cloud Run, Artifact Registry, Terraform |
| CI/CD | GitHub Actions, Workload Identity Federation |
