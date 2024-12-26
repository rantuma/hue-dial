# Contributing

Bug reports and pull requests are welcome.

## Architecture

The project follows hexagonal architecture — new behaviour belongs in `domain/` or `application/`; infrastructure adapters go in `infrastructure/`. Dependencies point inward: infrastructure depends on application and domain; application depends on domain only. Never import `infrastructure` or `application` from `domain`.

## Commit messages

This project follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification. Each commit message must have a structured format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Common types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.

A `feat` commit maps to a minor version bump; a `fix` commit maps to a patch bump. Append `!` after the type/scope or add a `BREAKING CHANGE:` footer to signal a breaking change (major bump).

Examples:

```
feat(hue): support multiple bridges
fix: handle dial event with no selected lamp
docs: update quick-start instructions
chore!: drop Go 1.21 support
```

## Running tests and linting

Run the full test and lint suite before submitting:

```bash
go test ./...
golangci-lint run
```
