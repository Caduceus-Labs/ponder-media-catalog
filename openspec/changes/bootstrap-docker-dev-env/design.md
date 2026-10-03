# Design

## Context

See `proposal.md` — Why. Current constraints that shape the approach:

- Go is not installed on the dev Mac (`go version` is absent). Docker 29.4 and
  Compose v2.34 are installed.
- The module was created once on a machine but never committed; `go.mod` must be
  regenerated and committed as part of this change.
- `docs/ARCHITECTURE.md` fixes the driver choice: `modernc.org/sqlite` is a
  pure-Go driver, so no C toolchain or CGO is required. Production runs a single
  static binary on a Linux VPS under systemd — Docker is a *local dev* concern only.
- The dev machine runs macOS on Apple silicon; the deploy target is Linux amd64.

## Goals / Non-Goals

**Goals:**
- One command set builds, tests, and runs the pipeline locally with no host Go.
- The same container image works on arm64 macOS and amd64 Linux.
- SQLite state persists on the host across container runs.
- Generated files (Go version pin, `go.mod`/`go.sum`) are committed so CI and
  future machines are reproducible.

**Non-Goals:**
- Containerizing production. The shipped artifact stays a bare binary + systemd timer.
- Devcontainer.json / VS Code integration (can be added later without changing this design).
- Multi-arch image *publication* (`docker buildx --push`). We only need native
  builds per platform for now.

## Decisions

### D1 — Compose service, not devcontainer, not raw `docker run`
Use a `compose.yaml` with one `dev` service. A `devcontainer.json` would tie the
workflow to VS Code; raw `docker run` invocations scattered in a README drift and
lose volume definitions. Compose keeps the volume/UID/platform settings in one
versioned file and is what `make` targets call. *Alternative considered:*
devcontainer (rejected — editor-coupled), `docker run` in Makefile (rejected —
env/volume duplication).

### D2 — Pin the official `golang` image; `CGO_ENABLED=0`
Base the image on the official `golang:1.25-bookworm` tag (record the resolved
digest in a comment) rather than `latest`, so `go vet`/tests are deterministic.
Because `modernc.org/sqlite` is pure Go, set `CGO_ENABLED=0`: the container needs
no gcc, and the same flag yields the static production binary. *Alternative:*
`golang:alpine` (rejected: musl edge cases with some tooling, larger first-build
friction for no benefit here).

### D3 — Bind-mounted source, named volumes for caches and data
Mount the repo as a bind mount (live edits, no rebuild). Put the Go build cache
and module cache on named volumes so container recreation does not re-download
dependencies. Put the SQLite directory on a separate named volume (`catalog-data`)
so the DB survives and is never inside the image. *Alternative:* copy source into
the image (rejected: kills the edit-run loop).

### D4 — Makefile is the stable entry point, wrapping the container
`make test`, `make vet`, `make build` run `docker compose run --rm dev …`. This
keeps the documented developer interface from `IMPLEMENTATION_PLAN.md` intact
while the implementation moves into the container, and lets CI reuse the same
targets. A `make shell` target drops into the container interactively.

### D5 — File ownership on macOS bind mounts
On Docker Desktop for Mac, bind-mounted files are owned by the host user but
written by the container's user. Fixed container UID 1000 vs the host user
(`hermes`/`calvink`, group `staff`) can produce `perm 600` files that VSCode
cannot open. Mitigation: run the `dev` service as the host UID/GID (passed via
`user:` from `id -u`/`id -g`) and set a permissive `umask`, so generated files
land `664`-readable. *Trade-off:* slightly more compose config; avoids the known
"can't open the file in the editor" failure.

### D6 — Commit `go.mod` / `go.sum` generated in the container
Regenerate the module inside the container (`go mod init` + `go get` the four
direct deps) and commit the result. This makes the container the single source of
truth for the dependency graph and removes the "it only exists on Calvin's
laptop" problem.

## Risks / Trade-offs

- **Bind-mount I/O and file-watch latency on macOS** → Acceptable for a
  build/test loop; if it bites, add a `cached`/`delegated` mount consistency hint.
- **Dependency download needs network on first run** → Module cache volume makes
  this a one-time cost; document `docker compose build` as the priming step.
- **UID/GID mismatch producing unreadable files (D5)** → Explicit `user:` mapping;
  verify by generating a file and checking mode is `664`.
- **`go.mod` claims `go 1.25.0` but the image tag drifts to a newer patch** →
  Pin the tag and record the digest; `go test` will surface any real mismatch.
- **Scope creep into Phase 1 code** → Non-goal; the entrypoint command handlers
  stay stubs that exit non-zero.

## Migration Plan

1. Add `.gitignore`, `Dockerfile.dev`, `compose.yaml`, Makefile.
2. `docker compose build` (primes toolchain + module cache).
3. Inside the container: `go mod init`, `go get` the four direct deps, commit
   `go.mod`/`go.sum`.
4. Create the directory layout + stub entrypoint; confirm `go build ./...` and
   `make test` pass inside the container.
5. Update README; tick Phase 0 checkboxes on issue #3.

Rollback: the change is additive (new files + a README edit); reverting the commit
restores the current state, as no existing code is modified.

## Open Questions

- Exact Go patch version to pin (1.25.x) — resolve at apply time from the image tag.
- Whether CI later runs the same `make` targets in the container — deferrable;
  no spec or task depends on it.
