# Spec Delta

## Purpose

Defines what a raw, source-originated catalog record is: the identity it must
carry and the content fingerprint used to detect change between ingestion runs.

## ADDED Requirements

### Requirement: Record Requires a Source Name

A record SHALL identify the source it came from. A record with an empty or
missing source name SHALL be rejected rather than constructed, so downstream
sync can never mistake an unattributed record for a valid one.

#### Scenario: Empty source name is rejected

- **WHEN** a record is constructed with an empty source name
- **THEN** construction fails with an error and no record is produced

#### Scenario: Non-empty source name is accepted

- **WHEN** a record is constructed with a non-empty source name
- **THEN** construction succeeds and the record carries that source name

### Requirement: Record Requires an External ID

A record SHALL carry the identifier the source assigns to it. A record with an
empty or missing external ID SHALL be rejected rather than constructed, so an
item can always be addressed and matched back to its origin.

#### Scenario: Empty external ID is rejected

- **WHEN** a record is constructed with an empty external ID
- **THEN** construction fails with an error and no record is produced

#### Scenario: Non-empty external ID is accepted

- **WHEN** a record is constructed with a non-empty external ID
- **THEN** construction succeeds and the record carries that external ID

### Requirement: Deterministic Content Fingerprint

A record SHALL expose a fingerprint derived deterministically from its raw
payload bytes, so that two records with identical payloads share a fingerprint
and any change to the payload yields a different one. The fingerprint SHALL be
computed from the payload alone.

#### Scenario: Identical payloads share a fingerprint

- **WHEN** two records are constructed from identical raw payloads
- **THEN** both records report the same fingerprint

#### Scenario: Different payloads differ in fingerprint

- **WHEN** two records are constructed from payloads that differ in any byte
- **THEN** the two records report different fingerprints

#### Scenario: Fingerprint depends only on the payload

- **WHEN** two records with different source names or external IDs are
  constructed from the same raw payload
- **THEN** both records report the same fingerprint

#### Scenario: Empty payload yields no fingerprint

- **WHEN** a record is constructed with an empty raw payload
- **THEN** the record reports an empty fingerprint
