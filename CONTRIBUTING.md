# Contributing to gowall

Thanks for your interest in contributing. This document covers how to get set up, the conventions we follow, and how to get a change merged.

## Getting started

```sh
git clone https://github.com/0mega24/gowall
cd gowall
go mod download
```

### Prerequisites

| Tool | Purpose |
|---|---|
| Go 1.26+ | Build and test |
| [golangci-lint](https://golangci-lint.run/welcome/install/) | Linting |
| [gofumpt](https://github.com/mvdan/gofumpt) | Formatting |

## Development workflow

```sh
make test      # go test -race -count=1 ./...
make vet       # go vet ./...
make lint      # golangci-lint run
make fmt       # gofumpt -extra -l -w .
make fix       # golangci-lint run --fix
```

All of these must pass before opening a PR. CI runs `vet`, `fmt`, and `lint` in the lint job and `test` in the test job.

## Making changes

1. Fork the repo and create a branch from `main`.
2. Branch names should follow the prefix convention: `feat/`, `fix/`, `chore/`, `docs/`.
3. Keep changes focused, one logical change per PR.
4. Add or update tests for any behavior change.
5. Ensure every new exported symbol has a doc comment and that any new package has a package-level doc comment (required by the `staticcheck` ST1000 rule).

## Commit messages

We follow the conventional commits style:

```
<type>: <short description>

<optional body>
```

Types: `feat`, `fix`, `chore`, `ci`, `docs`, `test`, `refactor`

- Use the imperative mood in the subject line ("add", not "adds" or "added")
- Keep the subject line under 72 characters
- Reference issues in the body where relevant

## Submitting a pull request

- Open a PR against `main`
- Fill in the PR template
- A PR should pass all CI checks before review
- We use **Rebase and merge**. Individual commits land on `main` as-is, so keep your history clean.

## Reporting issues

Please open a GitHub issue with enough detail to reproduce the problem: Go version, OS, wallpaper manager, and the actual vs expected output.
