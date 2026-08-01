# Ponder Catalog Pipeline

## Purpose

The Catalog Pipeline gets media data from external sources.
The pipeline changes source data into the Ponder catalog format.
The pipeline stores catalog data in a database.
The pipeline creates public JSON files for the Ponder application.
The pipeline publishes JSON files only when public catalog data changes.

## Scope

The Catalog Pipeline is separate from the Ponder application.
Ponder reads published JSON files.
The Catalog Pipeline does not provide a public API in the first release.
The Catalog Pipeline does not host the JSON files.

GitHub Pages hosts the public JSON files in the first release. GitHub Pages can publish static content from a GitHub repository.

## Main Requirements

- The pipeline must get data from multiple external sources.
- Each source must run as an independent job.
- A failed source job must not stop other source jobs.
- The pipeline must store source data incrementally.
- The pipeline must not replace the full catalog during each source job.
- The pipeline must detect changed source records.
- The pipeline must update only changed source records.
- The pipeline must create JSON files from canonical catalog data.
- The pipeline must not publish JSON files when their public content does not change.
- The pipeline must publish JSON files to a GitHub repository.
- GitHub Pages must serve the published JSON files.

## Non-Goals

The first release does not use:
- Airflow
- Microservices
- A message broker
- An event store
- Real-time source updates
- A public search API
- A cloud database

## System Structure

The system has four main parts:

```
External Source
    |
    v
Source Adapter
    |
    v
Catalog Database
    |
    v
Publisher
    |
    v
GitHub Pages JSON Files
    |
    v
Ponder Application
```

A source adapter gets data from one external source.
A source adapter changes source data into a source record.
The catalog module changes a source record into canonical catalog data.
The publisher reads canonical catalog data.
The publisher writes public JSON files.

## Deployment Model

- The system must use one Go binary.
- The binary must provide separate commands for each job.
- The system must not use one binary for each source.
- The system must not run as a long-running service in the first release.
- An external scheduler must start the commands.

Example commands:

```bash
ponder-catalog ingest source-a
ponder-catalog ingest source-b
ponder-catalog publish
ponder-catalog run-all
```

The `run-all` command is for local use, recovery, and initial catalog creation.
The `run-all` command must run source jobs in sequence.
The `run-all` command must publish only after source jobs complete.

## Scheduling

- The scheduler must start each source job independently.
- The scheduler must start the publisher on a separate schedule.
- The publisher schedule must start after expected source job completion.
- The scheduler must not start the publisher at the same time as a source job.

Example schedule:

```
06:10 UTC  ingest source-a
06:30 UTC  ingest source-b
07:00 UTC  publish
```

The first deployment should use a Linux systemd timer or cron.
A systemd timer can start a one-shot service at a specified time.
GitHub Actions can schedule commands later but is not the primary scheduler for a SQLite-based incremental catalog, because GitHub-hosted runners do not keep local files after a workflow ends.

## Source Adapter Requirements

- Each source adapter must implement the same source adapter interface.
- Each adapter must identify its source with a stable source name.
- Each adapter must get data from only one external source.
- Each adapter must map source data into source records.
- Each adapter must not expose external API models outside the adapter module.
- Each adapter must return stable external identifiers when the source provides them.
- Each adapter must record the source update time when the source provides it.
- Each adapter must save the raw source payload for debugging and later reprocessing.
- Each adapter must support retries for temporary network errors.
- Each adapter must use source rate limits.

## Catalog Requirements

- The catalog is the system of record for media data.
- The catalog must use a canonical media format.
- The canonical media format must not use source-specific field names.
- The catalog must store the relation between canonical media items and source records.
- The catalog must store source name and external source identifier.
- The catalog must support multiple source records for one media item.
- The catalog must keep source attribution for each value when required.
- The catalog must validate required media fields before publication.
- The catalog must not delete a media item when one source does not return it.
- The catalog must mark missing source records for later review.

## Database Requirements

The first release must use SQLite. SQLite is a self-contained database engine with no server process or configuration files.

- The database must persist between scheduled jobs.
- The database must use schema migrations.
- The database must contain these logical tables:

| Table | Purpose |
| :-- | :-- |
| `media_item` | Stores canonical media data |
| `source_record` | Stores data from one source |
| `media_source_link` | Links a media item to a source record |
| `source_sync_state` | Stores source checkpoints and job status |
| `publication` | Stores publication versions and file hashes |

- The `source_record` table must have a unique source name and external identifier pair.
- The database must store a fingerprint for each normalized source record.
- The pipeline must update a source record only when its fingerprint changes.
- The database must store the last successful source checkpoint.
- The pipeline must update a checkpoint only after a successful source job.

## Incremental Sync Requirements

- A source job must use source change data when the source provides it.
- A source job should use cursors, versions, ETags, or update timestamps.
- A source job may compare record fingerprints when change data is not available.
- A source job must be safe to run more than one time.
- A repeated job must not create duplicate catalog records.
- A failed job must not change the saved source checkpoint.
- A failed job must show the source name and failure reason in logs.

## Publishing Requirements

- The publisher must read canonical media data from the database.
- The publisher must create a public data transfer object.
- The publisher must not expose raw source payloads.
- The publisher must sort output records in a stable order.
- The publisher must serialize output in a stable format.
- The publisher must calculate a SHA-256 hash for each output file.
- The publisher must compare each new hash with the last published hash.
- If all hashes match, the publisher must exit without a commit.
- If any hash changes, the publisher must write changed JSON files.
- If any hash changes, the publisher must update the manifest file.
- If any hash changes, the publisher must commit and push output files.

## Public File Structure

The first release must use a manifest file.
The manifest file must list each public JSON file.
The manifest file must include a schema version.
The manifest file must include a catalog version.
The manifest file must include a file hash.

```
data/
  manifest.json
  media/
    movies.json
    series.json
    anime.json
    games.json
```

Example manifest:

```json
{
  "schemaVersion": 1,
  "catalogVersion": "2026-08-02T00:00:00Z",
  "files": [
    {
      "path": "media/anime.json",
      "sha256": "example-hash",
      "itemCount": 1234
    }
  ]
}
```

Ponder must get `manifest.json` before it gets catalog files.
Ponder must download a catalog file only when its hash changes.

## GitHub Pages Requirements

- The public data repository must contain only public JSON files and related metadata.
- The repository must not contain source credentials.
- The repository must not contain the SQLite database file.
- The repository must not contain raw source payloads.
- GitHub Pages must publish the repository content.
- The system must split catalog files before the output approaches GitHub Pages limits (1 GB repo, 100 GB/month bandwidth).

## Job Lock Requirements

- Each source job must use a source-specific lock.
- The publisher must use a publication lock.
- The system must not run two jobs for the same source at the same time.
- The system must not run two publisher jobs at the same time.
- The system may run different source jobs at the same time.
- The system must commit each source update in one database transaction.
- The publisher must read only committed catalog data.

## Domain Model

- `MediaItem` — canonical media data
- `SourceRecord` — data from one external source
- `SourceAdapter` — a source integration module
- `Publication` — one published catalog version
- `Catalog` — the canonical media collection

The source adapter module must contain external API models.
The application module must coordinate source sync and publication.
The domain module must contain canonical media rules.
The infrastructure module must contain database, HTTP, Git, and file system code.

## Domain Events

The first release may use in-process domain events:

- `SourceRecordChanged`
- `CatalogItemUpdated`
- `CatalogPublicationRequested`

The system must not require asynchronous event processing.
The system must not use Kafka, RabbitMQ, or an event store in the first release.
The system may replace in-process events with direct application service calls.

## Error Handling

- A source job must fail when it cannot validate required source data.
- A source job must not publish partial source data.
- A source job must log the source name for each error.
- A source job must log the external identifier when available.
- The publisher must fail when it cannot write a valid JSON file.
- The publisher must not push incomplete output files.
- The publisher must keep the last successful published data when a publication fails.

## Initial Delivery

1. Define the public media JSON schema.
2. Create the SQLite database schema.
3. Create one source adapter.
4. Add incremental source sync for the adapter.
5. Store canonical media data.
6. Create one JSON category file.
7. Create the manifest file.
8. Publish files to the GitHub Pages repository.
9. Make Ponder read the manifest file.
10. Add the next source adapter.