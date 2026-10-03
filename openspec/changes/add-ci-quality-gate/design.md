# Design

## Context

See `proposal.md` — Why. The relevant current state and constraints:

- The module targets Go 1.25 (`go.mod`); the local dev environment pins the same
  minor line as a digest-tagged container (`Dockerfile.dev`) and provides
  `make test`, `make test-race`, `make vet`, `make build` wrappers.
- The "no host Go toolchain" rule exists so developers do not install Go on their
  machines. It is a property of local development, not a requirement of the build.
- There is no `.github/` directory and no lint configuration today.
- Branch protection is repository state, not a committed file, so it cannot be
  expressed in the same artifact as the rest of the gate.

## Goals / Non-Goals

**Goals:**

- A single, fast, honest status check that runs on pull requests and main pushes.
- The check name is stable enough to require via branch protection.
- CI runs the same *toolchain version* as local development without sharing the
  same *packaging* (container).
- Lint rules are committed, so what the gate enforces is reviewable.

**Non-Goals:**

- Continuous delivery, release artifacts, or deploying the binary. Out of scope.
- Running the catalog `ingest`/`publish` jobs on Actions. The architecture keeps
  those on a persistent host with a durable SQLite file; Actions runners are
  ephemeral and unsuitable. See `docs/ARCHITECTURE.md` (Scheduling).
- Multi-version Go matrices. One pinned version is sufficient.
- Caching beyond the module cache that `actions/setup-go` provides by default.

## Decisions

### Run on the native runner, not the dev container

Use `actions/setup-go` with `go-version-file: go.mod`, then run the checks as
plain commands. **Why:** the Docker dev container exists to spare the developer a
host toolchain and to keep local environments identical; CI already has a clean,
ephemeral image and a toolchain one action away. Reusing the container would add
an image build to every run and couple CI to a dev-only concern.

*Alternatives:* (a) Reuse `docker compose run --rm dev …` for byte-for-byte parity
with local `make` targets — rejected for speed and because the parity it buys is
packaging, not behavior. (b) A hybrid that builds the image on a schedule — added
complexity for little value at this size. Revisit only if local/CI divergence ever
causes a real mismatch.

### One job, ordered steps

Express the whole gate as a single job named `ci` whose steps run in order. **Why:**
branch protection requires one entry per job; one job means one required check to
name and reason about, which matches the "single enforceable status check"
requirement.

*Alternatives:* Separate parallel jobs per concern (build/vet/lint/test) — nicer to
read, but produces several required checks that all have to be kept in sync, and
gains almost nothing while the module is this small.

### Lint from day one, with a committed config

Add `golangci-lint` now and commit its configuration file. **Why:** the choice was
made deliberately over deferring — installing the tool and its config while the
codebase is nearly empty establishes the habit and the file before there is real
code to retrofit. The config keeps the enforced scope explicit rather than hidden
in defaults.

*Alternatives:* Defer linting until a domain layer exists — fewer moving parts now,
but the tool and its config would be introduced later against a larger diff.

### Pin the toolchain from `go.mod`

Read the Go version from `go.mod` rather than duplicating it in the workflow.
**Why:** one source of truth; the Dockerfile digest and `go.mod` already stay in
step manually, and CI should add no third place to update.

### Enforcement is a repository setting, tracked as a task

The workflow file cannot require itself. Making the gate real is a branch-protection
change on `main` selecting the `ci` check. This is captured as an explicit task
because it lives outside version control and must not be forgotten.

### Pin golangci-lint to v2.14.0

The gate installs `golangci-lint` pinned to release `v2.14.0` (via the official
action at that tag), not a floating version. **Why:** go1.25 support arrived in
v2.4.0 (`🎉 go1.25 support`, released 2025-08-14), so the module's current
`go 1.25.0` pin is covered; v2.14.0 is the current release and is built with
go1.27, so it also covers the planned follow-up bump to Go 1.27 without another
CI change. Pinning a release — rather than `latest` — keeps CI and a developer's
local binary reporting the same findings.

*Alternatives:* (a) Track `latest` — drifts silently, and a new release can
introduce linter changes that turn the gate red without anyone touching the code.
(b) Pin the minimum that supports go1.25 (v2.4.0) — works today but sits at the
edge of the support window and would need bumping alongside the Go upgrade.

## Risks / Trade-offs

- **`golangci-lint` version drift between CI and a developer's local binary** →
  the action is pinned to release `v2.14.0` (see Decisions); document that same
  version for local use so "passes locally" and "passes in CI" agree.
- **The race detector needs cgo and a C compiler** → `ubuntu-latest` ships gcc and
  enables cgo by default, so `go test -race` works without extra setup. The dev
  container's default `CGO_ENABLED=0` is why local `make test-race` overrides it —
  the same reason applies here and needs no override.
- **Strict format and tidy checks add friction on first commit** → acceptable: both
  are trivially fixed by running `gofmt -w` and `go mod tidy`, and they prevent
  noise from accumulating.
- **Lint defaults may flag the std-lib-first style** → begin from the default linter
  set and tune the committed config only where a rule conflicts with the plan's
  conventions; do not disable rules wholesale.

## Migration Plan

1. Merge the workflow and lint config. On the first run the gate reports a status
   check on the pull request.
2. Once the check is observed passing, add it as a required status check on `main`.
3. Rollback: remove the required-check entry (merges unblock immediately) or delete
   the workflow file. Neither touches production code or the dev environment.
