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

- When a task spans multiple proposed commit boundaries, implement only one
  boundary at a time. After implementing and verifying that boundary, leave
  its changes unstaged, present them for user review, and stop before starting
  the next boundary.
- Do not accumulate changes for later commit boundaries in the working tree.
  When a file will be touched by multiple boundaries, change only the hunks
  required by the current boundary so each review and commit remains atomic.
- After the user approves and stages the current boundary, create an
  appropriate Conventional Commit without requesting confirmation again.
  Begin the next boundary only after that commit is complete.
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

## Terraform

### Running commands

Terraform is managed with mise. Run Terraform commands from the repository
root through `mise exec` and use `-chdir=infra`.

### Validation

After changing Terraform configuration, run:

```sh
mise exec -- terraform -chdir=infra fmt -check -diff
mise exec -- terraform -chdir=infra validate
git diff --check
```

### Style

- Group attributes by logical concern, such as iteration, resource identity,
  configuration, and lifecycle behavior. Insert a single blank line when the
  concern changes, but do not add blank lines mechanically at the start or end
  of a block.

### Safety

- Review deletion behavior for every Terraform-managed resource. When
  supported, explicitly protect resources whose deletion would remove
  artifacts, identities, or running services with
  `deletion_policy = "PREVENT"` or `deletion_protection = true`. Do not rely on
  a provider's default deletion behavior without deliberate review.
