# Tasks

## 1. Red — record identity validation

- [ ] 1.1 In a new `internal/catalog/domain/source_record_test.go`, write a test asserting that constructing a record with an empty source name fails; verify `docker compose run --rm dev go test ./internal/catalog/domain/ -run TestSourceRecord_RejectsEmptySourceName -v` fails to compile (`undefined: NewSourceRecord`).
- [ ] 1.2 In the same file, write a test asserting that constructing a record with an empty external ID fails; verify it too fails to compile until the constructor exists.

## 2. Green — implement identity and constructor

- [ ] 2.1 Implement the `SourceRecord` type and `NewSourceRecord` in `internal/catalog/domain/source_record.go`, rejecting an empty source name and an empty external ID; verify the tests from 1.1 and 1.2 now pass.
- [ ] 2.2 Delete the `internal/catalog/domain/.gitkeep` placeholder now that the package holds real files; verify `docker compose run --rm dev go build ./...` still succeeds.
- [ ] 2.3 Add doc comments to the exported type and constructor; verify `docker compose run --rm dev go doc ./internal/catalog/domain SourceRecord` renders them.

## 3. Red — content fingerprint behavior

- [ ] 3.1 Write a test asserting that two records built from identical payloads report the same fingerprint; verify it fails before the fingerprint is implemented.
- [ ] 3.2 Write a test asserting that two records built from payloads differing in any byte report different fingerprints; verify it fails.
- [ ] 3.3 Write a test asserting that a record built from an empty payload reports an empty fingerprint; verify it fails.

## 4. Green — implement the fingerprint

- [ ] 4.1 Implement the SHA-256 fingerprint over the raw payload bytes, returning an empty fingerprint for an empty payload; verify every test in `internal/catalog/domain/` passes (`docker compose run --rm dev go test ./internal/catalog/domain/ -v`).
- [ ] 4.2 Run the full local quality gate (`go build ./...`, `go vet ./...`, `gofmt -l .`, `go mod tidy` drift, `go test -race ./...`, lint) through the Docker dev env; verify it passes.

## 5. Integration & delivery

- [ ] 5.1 Confirm the domain package passes under the race detector (`docker compose run --rm dev go test -race ./internal/catalog/domain/`); verify no data race is reported.
- [ ] 5.2 Commit on a feature branch with message `feat: add SourceRecord domain value object with fingerprinting` and open a pull request; verify the required `ci` check passes on the pull request.
