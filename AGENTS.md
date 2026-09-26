# Development Instructions

## Workflow

### Before implementation

- When a task adds or changes directory or package boundaries, present the
  proposed structure before implementation, including the responsibilities and
  dependency direction of major additions. Do not create placeholders for
  components that have no current responsibility. Obtain user agreement on the
  structure before proceeding.
- Propose meaningful commit boundaries before implementation.

### Review and commits

- Leave changes unstaged for user review.
- After the user stages approved changes, create an appropriate Conventional
  Commit without requesting confirmation again.
- Use Conventional Commit types such as `feat` and `chore` for commits, and
  use the same type names as branch-name prefixes.

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
mkdir -p app/bin
mise exec -- go -C app build -o ./bin/lifnex ./cmd/lifnex
```

Do not rely on a globally installed Go version or invoke `go` directly.

### Package design

- Keep a helper local and unexported when it has a single consumer. Extract a
  package only when it represents a clear architectural boundary or has
  multiple consumers.
- Avoid unnecessary package proliferation under `internal` and avoid generic
  packages with unclear responsibilities.

### Testing

- Prefer an external `_test` package when testing only exported behavior. Use
  the same package when testing unexported implementation details is
  intentional.
- Use table-driven tests when cases share the same setup, action, and assertion
  structure. Keep behaviorally distinct scenarios in separate test functions.

### Style

- **Logical spacing:** Keep closely related statements together, and insert a
  single blank line when the code moves to a distinct logical step, concern, or
  phase. Apply this to control flow and `return` statements based on meaning,
  rather than adding blank lines mechanically. Do not add blank lines solely
  because a function or control-flow block starts, a `case` or `default` label
  appears, or a block ends.

## MCP

### Testing

- Test tool registration separately from the behavior of each tool.

### Stdio transport

- Reserve `stdout` exclusively for MCP protocol messages when using stdio
  transport. Write application logs to `stderr`.
