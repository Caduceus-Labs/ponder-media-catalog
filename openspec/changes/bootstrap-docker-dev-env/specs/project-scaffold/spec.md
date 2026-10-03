# Spec Delta

## Purpose

Defines the repository's Go module layout, its single-binary build, and the
build/test tooling every later phase of the catalog pipeline depends on.

## ADDED Requirements

### Requirement: Single Go Module

The repository SHALL contain exactly one Go module, named
`github.com/Caduceus-Labs/ponder-media-catalog`, whose `go.mod` and `go.sum`
are committed to version control.

#### Scenario: Module resolves from a clean checkout

- **WHEN** a developer runs `go build ./...` on a fresh clone
- **THEN** the module resolves without fetching undeclared modules
- **AND** all import paths are rooted at `github.com/Caduceus-Labs/ponder-media-catalog`

#### Scenario: Dependencies are declared

- **WHEN** the module is inspected
- **THEN** the direct dependencies are the CLI parser, the pure-Go SQLite driver,
  the migration library, and the test assertion library named in the implementation plan
- **AND** no dependency beyond that set is required to build

### Requirement: Single Binary Entrypoint

The system SHALL build as one binary that dispatches to named subcommands, and
SHALL NOT build one binary per source or run as a long-running service.

#### Scenario: Build succeeds

- **WHEN** a developer builds the entrypoint package
- **THEN** compilation succeeds and a single executable is produced

#### Scenario: Command surface is declared

- **WHEN** the binary is invoked with no arguments or an unknown command
- **THEN** it lists the declared commands `ingest`, `publish`, and `run-all`
- **AND** exits non-zero for an unknown command

#### Scenario: Unimplemented commands do not pretend to succeed

- **WHEN** a declared but not-yet-implemented command is invoked
- **THEN** it exits with a non-zero status indicating it is not implemented
- **AND** it does not write to the database or the output files

### Requirement: Documented Directory Layout

The repository SHALL use the directory layout defined in the implementation plan,
separating the CLI, the catalog bounded context (domain, application,
infrastructure), shared platform packages, migrations, and test data.

#### Scenario: Layout matches the plan

- **WHEN** the repository tree is listed
- **THEN** it contains `cmd/ponder-catalog`, `internal/catalog/domain`,
  `internal/catalog/application`, `internal/catalog/infrastructure`,
  `internal/platform`, `internal/cli`, `migrations`, and `testdata`
- **AND** domain code depends on no infrastructure package

### Requirement: Build and Test Tooling

The repository SHALL provide a Makefile exposing at least `test`, `test-race`,
`vet`, `build`, and `clean` targets that operate on the whole module.

#### Scenario: Test target runs the suite

- **WHEN** a developer runs `make test`
- **THEN** the module's Go tests run and the target fails if any test fails

#### Scenario: Build target produces the binary

- **WHEN** a developer runs `make build`
- **THEN** the single binary is written under `bin/`

#### Scenario: Clean target removes build output

- **WHEN** a developer runs `make clean`
- **THEN** the `bin/` build output is removed

### Requirement: Repository Hygiene

The repository SHALL ignore build output, local database files, and local
environment files, and SHALL carry a README describing the project.

#### Scenario: Build artifacts are not tracked

- **WHEN** the binary has been built locally
- **THEN** the build output directory and `*.db` / `*.sqlite` files do not appear
  as untracked changes

#### Scenario: README describes the project

- **WHEN** the README is read
- **THEN** it names the project and states that it produces the public catalog
  JSON consumed by the Ponder application
