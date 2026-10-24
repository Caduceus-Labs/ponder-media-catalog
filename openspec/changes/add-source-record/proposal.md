# Proposal

## Why

The catalog pipeline ingests raw records from external media sources and must
detect when a record's content changes between runs. Today the repository has a
compiling module (Phase 0) but no domain logic at all — nothing defines what a
raw source record *is*, and nothing can tell whether two records are equivalent.

`SourceRecord` is the smallest unit of domain logic with real business value: it
validates that a record carries the identity it needs, and it computes a
fingerprint of its raw payload so downstream sync can tell "unchanged" from
"changed" without diffing payloads field by field. It has no database, no HTTP,
and no CLI — it is pure rules, which makes it the right first piece of the domain
layer and the foundation the `SyncSource` use case (Phase 4) will build on.

## What Changes

- Introduce a `SourceRecord` domain value object representing one raw record
  pulled from a named source.
- Enforce two invariants at construction: a non-empty source name and a non-empty
  external ID. Construction fails otherwise.
- Compute a deterministic SHA-256 fingerprint of the record's raw bytes at
  construction, so identical data yields an identical fingerprint and any byte
  change yields a different one.
- Cover all of the above with table-driven unit tests, written test-first.

Out of scope: persistence, change detection *policy* (acting on a fingerprint
difference), the source adapters that produce records, and the canonical
`MediaItem` model (Phase 2).

## Capabilities

### New Capabilities
- `source-record`: the raw, source-originated record value object — its required
  identity fields, its validation rules, and its content fingerprint used for
  change detection.

### Modified Capabilities
- None. This introduces a new capability; no existing spec's requirements change.

## Impact

- **Repo**: adds `internal/catalog/domain/source_record.go` and
  `internal/catalog/domain/source_record_test.go`. The `domain` package becomes
  real (its `.gitkeep` placeholder is removed).
- **Dependencies**: standard library only (`crypto/sha256`, `fmt`, `testing`).
  No new Go modules.
- **Behavior**: purely additive — no existing code path calls the new type yet.
  The existing CI quality gate (`build`/`vet`/`gofmt`/`tidy`/`test -race`/`lint`)
  automatically covers the new package.
- **Downstream**: unblocks Phase 4 (`SyncSource`), which consumes `SourceRecord`
  fingerprints to decide whether a source item changed.
