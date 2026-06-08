# Contributing to Roket.Chat-GeekBot

Thank you for considering contributing! This document outlines the process for contributing code, reporting issues, and submitting changes.

## Quick Start

1. Fork the repository
2. Open in GitHub Codespaces (or clone locally with Go 1.22+)
3. Run `go mod download`
4. Copy `.env.example` to `.env` and configure your Rocket.Chat server
5. Run `go run ./cmd/bot`
6. Run `go test ./...` to verify everything works

## Branch Strategy

This project uses a **feature flow** branching model:

```
main (stable, protected)
  ├── feat/rocket-connection     ← new feature
  ├── feat/standup-report        ← new feature
  ├── fix/login-timeout          ← bug fix
  ├── docs/contributing          ← documentation
  └── chore/update-deps          ← maintenance
```

| Branch prefix | Purpose |
|---------------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation changes |
| `refactor/` | Code refactoring |
| `test/` | Adding or updating tests |
| `chore/` | Maintenance, dependencies, config |

All branches are created from `main` and merged back via pull requests.

## Commit Convention

This project uses **Conventional Commits** for all commit messages.

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to use |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only |
| `style` | Formatting, linting (no logic change) |
| `refactor` | Code restructuring (no feature/fix) |
| `perf` | Performance improvement |
| `test` | Adding or updating tests |
| `build` | Build system or dependencies |
| `ci` | CI configuration changes |
| `chore` | Maintenance, tooling, config |

### Examples

```
feat: add standup report generation

Implement /standup report command that aggregates
all responses and posts formatted output to channel.

Closes #8
```

```
fix: handle empty standup submissions gracefully

Prevent panic when user submits with no answers.

Fixes #12
```

```
docs: add API reference for slash commands
```

### Rules

- Use present tense, imperative mood: "add" not "added" or "adds"
- Keep the first line under 72 characters
- Reference issues in the body or footer: `Closes #123`, `Refs #456`
- One logical change per commit

## Pull Request Process

1. Create a feature branch from `main`: `git checkout -b feat/my-feature`
2. Make your changes with conventional commit messages
3. Ensure all tests pass: `go test ./...`
4. Ensure code is clean: `go vet ./...`
5. Push your branch and open a PR against `main`
6. In the PR description, explain what changed and why
7. Link related issues (e.g., "Closes #4")
8. Wait for review and address any feedback

### PR Title Format

Use the same conventional commit format:
```
feat: add team admin CRUD commands
fix: prevent duplicate standup submissions
docs: update command reference table
```

## Development Guidelines

### Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` — zero warnings required
- Follow standard Go conventions (effective Go, Go idioms)
- Use meaningful names — no `temp`, `data`, `result` without context
- Prefer standard library over external dependencies

### Testing

- Write tests for all new functionality
- Run tests: `go test ./...`
- Aim for meaningful test coverage (not just happy paths)
- Test edge cases: empty inputs, unauthorized users, duplicate submissions

### Adding a New Command

1. Create a handler in `internal/commands/`
2. Register it in the command registry
3. Add permission checks in the middleware
4. Write tests for the handler
5. Update the README command table

### Dependencies

Before adding a new dependency:
1. Check if the existing code solves the problem
2. Check the dependency's maintenance status and license
3. Run `go mod tidy` after adding

## Reporting Issues

### Bug Reports

When filing a bug, include:
- Steps to reproduce
- Expected vs actual behavior
- Bot logs (if available)
- Rocket.Chat server version

### Feature Requests

Describe:
- What you want to accomplish
- Why it's useful
- How you envision it working

## Code of Conduct

Be respectful, constructive, and inclusive. Disagreements are fine — personal attacks are not.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
