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
