# Spec Delta

## Purpose

Provides an automated quality gate that validates every proposed change to the
module — building it, analyzing it statically, running its tests with the race
detector, and linting it — so that regressions are caught before they reach the
main branch.

## ADDED Requirements

### Requirement: Gate Runs on Pull Requests and Main

The repository SHALL run its quality gate automatically whenever a pull request
is opened or updated and whenever a commit is pushed to the main branch, without
requiring a manual trigger.

#### Scenario: Pull request triggers the gate

- **WHEN** a pull request against the main branch is opened or updated
- **THEN** the quality gate runs against the pull request's head commit
- **AND** its result is reported on the pull request

#### Scenario: Main branch push triggers the gate

- **WHEN** a commit is pushed to the main branch
- **THEN** the quality gate runs against that commit

### Requirement: Toolchain Version Alignment

The gate SHALL use the Go toolchain version declared by the module rather than a
fixed or floating CI version, so the version the gate tests matches the version
the module targets.

#### Scenario: CI version follows the module

- **WHEN** the Go version declared in `go.mod` changes
- **THEN** the gate's build and test steps run under the newly declared version
- **AND** no separate CI configuration value needs editing to keep them aligned

### Requirement: Build Verification

The gate SHALL compile the entire module and SHALL fail the check if any package
fails to build.

#### Scenario: Compilation failure fails the gate

- **WHEN** the module contains a package that does not compile
- **THEN** the gate reports a failed check

### Requirement: Static Analysis

The gate SHALL run `go vet` across the module and SHALL fail the check when vet
reports findings.

#### Scenario: Vet finding fails the gate

- **WHEN** `go vet` reports a finding in any package
- **THEN** the gate reports a failed check

### Requirement: Formatting Enforcement

The gate SHALL verify that every Go file is `gofmt`-clean and SHALL fail the
check when any file is not formatted.

#### Scenario: Unformatted source fails the gate

- **WHEN** a committed Go file is not formatted according to `gofmt`
- **THEN** the gate reports a failed check

### Requirement: Dependency Tidiness

The gate SHALL verify that `go.mod` and `go.sum` are consistent with the module's
imports and SHALL fail the check when they are not tidy.

#### Scenario: Stale module files fail the gate

- **WHEN** `go.mod` or `go.sum` does not match what the module's imports require
- **THEN** the gate reports a failed check

### Requirement: Race-Enabled Test Execution

The gate SHALL run the module's test suite with Go's race detector enabled and
SHALL fail the check when any test fails or the race detector reports a data race.

#### Scenario: Test failure fails the gate

- **WHEN** any test in the module fails
- **THEN** the gate reports a failed check

#### Scenario: Data race fails the gate

- **WHEN** the race detector reports a data race during the test run
- **THEN** the gate reports a failed check

#### Scenario: Passing suite passes the gate

- **WHEN** the module builds and every test passes without a reported race
- **THEN** the gate reports a successful check

### Requirement: Configured Linting

The gate SHALL run a linter configured by a committed configuration file, and
SHALL fail the check when the linter reports a finding within the configured
scope. The lint configuration SHALL live in the repository so the enforced rules
are explicit and version-controlled.

#### Scenario: Lint finding fails the gate

- **WHEN** the linter reports a finding within the configured scope
- **THEN** the gate reports a failed check

#### Scenario: Lint scope is version-controlled

- **WHEN** the repository is cloned
- **THEN** the lint configuration is present and the gate applies it without
  additional setup

### Requirement: Single Enforceable Status Check

The gate SHALL report as a single named status check, so the main branch can
require that one check before a pull request is merged.

#### Scenario: Gate is selectable as a required check

- **WHEN** branch protection for the main branch is configured
- **THEN** the gate's status check is available to select as a required check

#### Scenario: Required gate blocks merge while failing

- **WHEN** the gate is required for the main branch and reports a failed check on
  a pull request
- **THEN** the pull request cannot be merged until the check passes

### Requirement: Immutable Gate Definition

The gate's workflow definition SHALL be committed to the repository so that any
change to what the gate checks passes through the same pull request review as the
code it validates.

#### Scenario: Gate definition is version-controlled

- **WHEN** the repository is cloned
- **THEN** the workflow definition is present under `.github/workflows/`
- **AND** no gate step depends on configuration that exists only in repository
  settings or a web UI
