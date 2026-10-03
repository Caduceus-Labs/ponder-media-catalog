# Proposal

## Why

Phase 0 (Project Bootstrap, issue #3) is only half done: `go mod init` and `go get`
ran once on Calvin's machine but nothing was committed — the repo still has no
`go.mod`, no directory layout, no entrypoint, and no Makefile. On top of that, Go
is not installed on the dev Mac, so "run the tests / build the binary" is not
reproducible on the machine where development happens.

This change finishes Phase 0 and makes the toolchain portable by running all Go
work inside Docker. That unblocks every later phase (domain → application →
infrastructure), which all assume a compiling module.

## What Changes

- Commit a `go.mod` / `go.sum` for `github.com/Caduceus-Labs/ponder-media-catalog`
  with the four direct dependencies from the plan (cobra, modernc sqlite,
  golang-migrate, testify). The dependency set is already sufficient for the
  bootstrap — no new libraries are required.
- Create the directory layout from `docs/IMPLEMENTATION_PLAN.md` Phase 0.3
  (`cmd/ponder-catalog`, `internal/catalog/{domain,application,infrastructure/*}`,
  `internal/platform/*`, `internal/cli`, `migrations`, `testdata/`).
- Add a minimal `cmd/ponder-catalog/main.go` whose only job is to compile and
  dispatch to the CLI package, so `go build ./...` succeeds and `make build`
  produces `bin/ponder-catalog`.
- Add a rooted CLI command surface (cobra) with the four commands named in
  `docs/ARCHITECTURE.md` (`ingest <source>`, `publish`, `run-all`) as declared,
  not-yet-implemented stubs.
- Add a `Makefile` with `test`, `test-race`, `vet`, `build`, `clean` targets.
- Replace the stale `media-scraper` README with the project one-liner and add a
  `.gitignore` covering `bin/`, `*.db`, `*.sqlite`, and local env files.
- Add a Docker dev environment: a `Dockerfile.dev` pinning the Go toolchain and a
  `compose.yaml` exposing build / test / vet / shell targets, so no host Go
  install is needed. The SQLite database lives on a mounted volume, not in the
  image.
- Mark the completed Phase 0 checkboxes in issue #3.

Out of scope: any domain/application/infrastructure code (Phases 1+), migrations,
the GitHub Pages data repo, and VPS provisioning.

## Capabilities

### New Capabilities
- `project-scaffold`: the repository's Go module layout, its single-binary build,
  and the build/test tooling (Makefile, ignore rules) that every later phase builds on.
- `docker-dev-environment`: a containerized, host-independent Go toolchain used to
  build, test, and run the pipeline binary locally.

### Modified Capabilities
- None. `openspec/specs/` is empty; this is the first change to introduce specs.

## Impact

- **Repo**: adds `go.mod`, `go.sum`, `cmd/`, `internal/`, `migrations/`,
  `testdata/`, `Makefile`, `.gitignore`, `Dockerfile.dev`, `compose.yaml`; updates
  `README.md`; adds two specs.
- **Dependencies**: no new Go modules beyond the plan's four (plus their
  transitive entries). No runtime services — SQLite stays a single file.
- **Issues**: completes Phase 0 (#3) and unblocks the `Go project scaffolded`
  deliverable of Epic 1 (#1). No issue is closable until this change is applied
  and artifacts are committed.
- **Deployment**: unaffected — production remains a single binary on a Linux VPS
  under a systemd timer; Docker is a local dev convenience only.
