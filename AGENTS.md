# Development Instructions

## Go

Go is managed with mise. Run Go commands from the repository root through
`mise exec` so that non-interactive shells and coding agents use the version
declared in `mise.toml`.

For first-time setup:

```sh
mise trust
mise install
```

For daily development:

```sh
mise exec -- go -C app test ./...
mise exec -- go -C app build ./cmd/lifnex
```

Do not rely on a globally installed Go version or invoke `go` directly.
