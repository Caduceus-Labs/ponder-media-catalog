# Proposal

## Why

The repository has no automated quality gate. Every check — build, `go vet`,
formatting, dependency tidiness, tests, lint — depends on a developer remembering
to run it locally before merging, so a regression can reach `main` unseen. Now is
the cheapest moment to close that gap: the module is still small (one declared CLI
surface, no tests yet), so the gate can land before there is any backlog to
retrofit it across, and every test written from Phase 1 onward is covered from the
moment it is committed.

## What Changes

- Add a GitHub Actions workflow that runs on pull requests and on pushes to `main`.
- The workflow runs one quality gate job: build, `go vet`, `gofmt` cleanliness
  check, `go mod tidy` drift check, `go test -race ./...`, and `golangci-lint run`.
- Pin the Go toolchain to the version declared in `go.mod` via
  `actions/setup-go` (`go-version-file: go.mod`), keeping CI aligned with the
  pinned Docker dev toolchain.
- Add a committed `golangci-lint` configuration file so the lint scope is explicit
  and version-controlled from day one.
- Run the gate on the native GitHub runner toolchain rather than the Docker dev
  container: the "no host Go" rule is a local-development constraint, not a CI
  requirement.
- Expose the gate as a single acceptably-named status check so branch protection
  on `main` can require it and block merges while the gate is red.

No breaking changes. No change to production code or the development environment.

## Capabilities

### New Capabilities

- `continuous-integration`: an automated build, static-analysis, test, and lint
  gate that runs on every pull request and push to the main branch, reports a
  single pass/fail check, and is enforced as a required status check before merge.

### Modified Capabilities

_None. No existing capability's requirements change: `project-scaffold` and
`docker-dev-environment` describe local build/test tooling, which this change
leaves intact._

## Impact

- **New files**: `.github/workflows/ci.yml` (the gate), `.golangci.yml` (lint
  configuration).
- **Dependencies**: none added to `go.mod`; `golangci-lint` is a CI-only tool
  installed by the workflow, not a module dependency.
- **Code**: no production, domain, or infrastructure code changes.
- **Repository settings (outside version control)**: branch protection for `main`
  must be updated to require the gate's status check. Captured as an explicit task
  since it cannot be expressed as a committed file.
- **Out of scope (follow-up)**: the module's `go 1.25.0` pin is now outside the
  two-latest-minor support windows of both the Go toolchain and golangci-lint.
  Bringing the toolchain current (for example to Go 1.27) is deliberately excluded
  from this change and should land as a separate follow-up, validated by the gate
  this change introduces.
