# Ponder Media Catalog

Periodic catalog pipeline that ingests media metadata from external sources,
stores it incrementally in SQLite, and publishes the public JSON files that the
[Ponder](https://github.com/Caduceus-Labs) application reads.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full design and
[`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md) for the phased build plan.

## Development

No host Go toolchain is required — everything runs in the Docker dev container.

```bash
docker compose build   # one-time: prime the toolchain and module cache
make test              # run the Go test suite
make build             # build bin/ponder-catalog
make shell             # interactive shell with the toolchain
```

`make` targets wrap `docker compose run --rm dev …` (see `Dockerfile.dev` and
`compose.yaml`).

## Continuous integration

Every pull request and every push to `main` runs a single quality gate
(`.github/workflows/ci.yml`, job `ci`). The job is intended to be configured as a
required status check on `main`, so a red run blocks merge. It runs on the
runner's Go toolchain using the version declared in `go.mod` — no Docker is
involved in CI.

The gate runs these checks:

| Check | Command |
| :-- | :-- |
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Formatting | `gofmt -l .` must print nothing |
| Module tidiness | `go mod tidy`, then `git diff --exit-code -- go.mod go.sum` |
| Tests (race detector) | `go test -race ./...` |
| Lint | `golangci-lint run` (configuration in `.golangci.yml`) |

### Reproducing the gate locally

Each check maps to a `make` target that runs it the same way CI does:

```bash
make build       # go build ./...
make vet         # go vet ./...
make fmt-check   # gofmt -l . must print nothing
make tidy-check  # go mod tidy, then git diff --exit-code -- go.mod go.sum
make test-race   # go test -race ./...
make lint        # golangci-lint run
```

`golangci-lint` is pinned to **v2.14.0** (see `.golangci.yml` and the workflow) so
a local run and a CI run report identical findings.
