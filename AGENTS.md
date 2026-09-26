# Development Instructions

## Go

### Running commands

Go is managed with mise. Run Go commands from the repository root through
`mise exec` so that non-interactive shells and coding agents use the version
declared in `mise.toml`.

**First-time setup:**

```sh
mise trust
mise install
```

**Daily development:**

```sh
mise exec -- go -C app test ./...
mise exec -- go -C app build ./cmd/lifnex
```

Do not rely on a globally installed Go version or invoke `go` directly.

### Style

- **Logical spacing:** Keep closely related statements together, and insert a
  single blank line when the code moves to a distinct logical step, concern, or
  phase. Apply this to control flow and `return` statements based on meaning,
  rather than adding blank lines mechanically. Do not add blank lines solely
  because a function or control-flow block starts, a `case` or `default` label
  appears, or a block ends.
