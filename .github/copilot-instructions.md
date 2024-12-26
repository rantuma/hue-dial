# GitHub Copilot Instructions

## Project Overview

This is a Go service that reacts to Philips Hue button and dial events and controls Hue lights accordingly. It reads device mappings from `config.json` and bridges physical button presses/dial turns to lamp on/off and brightness actions.

## Architecture

The project follows hexagonal (ports and adapters) architecture:

- `domain/` — Core types with no external dependencies (`device`, `config`, `ports`)
- `application/` — Application services orchestrating domain logic (`event.Service`)
- `infrastructure/` — Concrete adapters (`hue`, `config`, `logging`, `setup`)
- `pkg/` — Shared utilities (`logger`, `version`)

Dependencies point inward: infrastructure depends on application and domain; application depends on domain only. Never import `infrastructure` or `application` from `domain`.

Ports (interfaces) are defined in `domain/ports` and implemented by infrastructure adapters. The application layer depends only on these interfaces, never on concrete types.

## Code Conventions

- Constructor functions are always named `New(...)` and return an interface or pointer plus an `error`.
- Wrap errors with context: `fmt.Errorf("failed to do X: %w", err)`.
- Reuse the `err` variable name with `=` assignment instead of inventing names like `mkdirErr` or `unmarshalErr` to avoid shadowing. Use `//nolint:govet // shadow` if the linter complains about `err` reuse.
- Use `log.Panicf(...)` at the top-level `main` for unrecoverable startup errors.
- Suppress linter warnings with `//nolint:<linter> // <justification>` — always include a reason.
- Package-level variables are avoided; use `//nolint:gochecknoglobals` with justification when unavoidable (e.g., config singleton).
- Config is stored as JSON at `/data/config.json` (overridable via `CONFIG_PATH` env var) and loaded with `encoding/json`. The initial config is produced by an interactive TUI setup wizard (`infrastructure/setup`) built with `charm.land/huh/v2`.

## Testing

- Use `github.com/stretchr/testify/assert` (not `require`) for assertions.
- Tests live in external test packages (`package foo_test`).
- Use table-driven tests with a `tests []struct{ name, ... }` slice and `t.Run(tt.name, ...)`.
- Mock ports by implementing the interface directly in the test file.

## Build & Validation

After making changes, always run:

```sh
make test
make lint
```

After adding or updating dependencies, regenerate license files:

```sh
make licenses
```

Build the binary with version injection:

```sh
make build
```


## Comments

Do not add comments unless they are strictly required to understand the code — i.e. the code alone cannot convey the intent. Never add comments that merely restate what the function name, type name, or variable name already says.
