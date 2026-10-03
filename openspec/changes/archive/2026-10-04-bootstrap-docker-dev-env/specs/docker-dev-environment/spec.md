# Spec Delta

## Purpose

Provides a containerized Go toolchain so the catalog pipeline can be built,
tested, and run locally without a host Go installation, keeping the development
environment identical across machines and operating systems.

## ADDED Requirements

### Requirement: Pinned Container Toolchain

The repository SHALL define a container image that provides the Go toolchain
required by the module, pinned to a specific version rather than a floating tag.

#### Scenario: Image builds reproducibly

- **WHEN** the image is built from its definition
- **THEN** it installs the same Go version and module dependencies every time
- **AND** it does not depend on a `latest` toolchain tag

#### Scenario: Host Go is not required

- **WHEN** a developer with no Go installed runs the containerized build
- **THEN** the build completes using only the toolchain inside the image

### Requirement: Containerized Build, Test, and Run

The development environment SHALL expose build, test, vet, and interactive shell
operations through container orchestration, so the pipeline can be exercised
without host setup.

#### Scenario: Tests run in the container

- **WHEN** a developer runs the containerized test task
- **THEN** the module's Go tests execute inside the container
- **AND** the task's exit status reflects the test result

#### Scenario: Build runs in the container

- **WHEN** a developer runs the containerized build task
- **THEN** the single binary is produced as it would be on a host

#### Scenario: Interactive shell is available

- **WHEN** a developer starts the containerized shell
- **THEN** they get a shell with the Go toolchain and the module mounted for editing

### Requirement: Persistent Local Data

Database files produced during local development SHALL live on a mounted volume
outside the container image, so data survives container restarts and is never
baked into an image.

#### Scenario: Database survives a container run

- **WHEN** a command writes the SQLite database and the container is later recreated
- **THEN** the database file is still present at the same host-visible path

#### Scenario: Image carries no data

- **WHEN** the built image is inspected
- **THEN** it contains no database file, no build output, and no source secrets

### Requirement: Cross-Platform Development

The development environment SHALL work unchanged on the developers' supported
host platforms, including Apple silicon macOS and Linux.

#### Scenario: Runs on Apple silicon and Linux

- **WHEN** the container is started on Apple silicon macOS or on Linux
- **THEN** the toolchain executes natively for the host architecture
- **AND** the same build and test tasks succeed on both
