# Tasks

## 1. Container Toolchain

- [x] 1.1 Add `Dockerfile.dev` pinning the Go toolchain (`golang:1.25-bookworm`, digest recorded in a comment) with `CGO_ENABLED=0`; verify with `docker compose build` succeeding on Apple silicon.
- [x] 1.2 Add `compose.yaml` defining a `dev` service that bind-mounts the repo, runs as the host UID/GID, and declares named volumes for the Go module cache, build cache, and `catalog-data`; verify `docker compose config` resolves without warnings.
- [x] 1.3 Verify the container has no host dependency: run `docker compose run --rm dev go version` and confirm it prints the pinned version; confirm the image contains no database file or build output.

## 2. Module Bootstrap (inside the container)

- [x] 2.1 Run `go mod init github.com/Caduceus-Labs/ponder-media-catalog` in the container and verify `go.mod` appears with the correct module path.
- [x] 2.2 `go get` the four direct dependencies (cobra, `modernc.org/sqlite`, `golang-migrate/migrate/v4`, `testify`) and verify `go.mod`/`go.sum` list them and `go mod tidy` reports no changes.
- [ ] 2.3 Commit `go.mod` and `go.sum` and verify a clean checkout resolves with `go build ./...` returning success.

## 3. Layout and Entrypoint

- [x] 3.1 Create the directory layout from Phase 0.3 (`cmd/ponder-catalog`, `internal/catalog/{domain,application,infrastructure}`, `internal/platform`, `internal/cli`, `migrations`, `testdata`) and verify the tree matches the plan.
- [x] 3.2 Add `cmd/ponder-catalog/main.go` that delegates to the CLI package; verify `go build -o bin/ponder-catalog ./cmd/ponder-catalog/` produces a single binary.
- [x] 3.3 Add the cobra command surface declaring `ingest`, `publish`, and `run-all`; verify an unknown command exits non-zero and the declared commands are listed.
- [x] 3.4 Make each declared-but-unimplemented command exit non-zero without touching the database or output; verify by running `ingest`/`publish`/`run-all` and checking exit status and that no files were written.

## 4. Build Tooling and Repository Hygiene

- [x] 4.1 Add a Makefile with `test`, `test-race`, `vet`, `build`, `clean`, and `shell` targets that wrap `docker compose run --rm dev`; verify `make build` writes `bin/ponder-catalog` and `make clean` removes it.
- [x] 4.2 Add `.gitignore` for `bin/`, `*.db`, `*.sqlite`, and local env files; verify that a built binary and a created `*.db` do not show as untracked changes.
- [x] 4.3 Replace the README with the project name and a one-line statement that it produces the public catalog JSON consumed by Ponder; verify the statement reads correctly.

## 5. Integration Verification

- [x] 5.1 From a clean checkout with no host Go, run `make vet`, `make test`, and `make build` and verify all succeed using only the container.
- [x] 5.2 Verify database persistence: create a file under the `catalog-data` volume path, recreate the container, and confirm the file is still present.
- [x] 5.3 Verify file ownership: generate a file in the bind mount and confirm its mode is group- and world-readable (`664`).
- [ ] 5.4 Tick the completed Phase 0 checkboxes (0.3–0.5) on issue #3 and note the Docker dev environment.
