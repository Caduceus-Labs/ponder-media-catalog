# Design

## Context

See `proposal.md` — Why. The repository currently has no domain package content;
`internal/catalog/domain/` holds only a `.gitkeep`. The module builds and is
gated by CI, but nothing has been written against it yet. This design covers how
the first domain type is shaped. Requirements are in
`specs/source-record/spec.md`.

## Goals / Non-Goals

**Goals:**
- A single, dependency-free value type that enforces its own invariants.
- A fingerprint that is cheap to compute and stable for a given payload.

**Non-Goals:**
- Storing, comparing, or acting on fingerprints across runs (that is `SyncSource`, Phase 4).
- Any source-specific parsing or normalization of payloads.
- Freezing the type's final shape — later phases may extend it.

## Decisions

**1. Construct through a validating constructor, not a bare struct literal.**
`NewSourceRecord(...) (*SourceRecord, error)` returns an error when a required
field is empty, so an invalid record cannot exist. *Alternative considered:* a
plain exported struct plus a separate `Validate()` method — rejected, because it
permits invalid values to be created and only catches them if every caller
remembers to validate. A constructor makes the invariant structural.

**2. Fingerprint is SHA-256 over the raw payload bytes.**
The sources deliver an opaque payload per record; the only question downstream
asks is "did these bytes change?". Hashing the bytes answers exactly that without
the domain layer knowing the payload's schema. *Alternative considered:*
canonicalize the payload (e.g. parse JSON, sort keys, re-serialize, then hash) —
rejected as premature: it couples the domain to a payload format the domain does
not otherwise care about, and can be introduced later behind the same signature
if a source proves noisy.

**3. The fingerprint deliberately excludes source name and external ID.**
Identity answers *which item this is*; the fingerprint answers *whether its
content changed*. Keeping them orthogonal means the same payload under two
identities fingerprints identically, which is what makes equality checks
meaningful. *Alternative considered:* folding source + ID into the hash —
rejected, as it would make two identical payloads appear "changed" merely because
they are addressed differently.

**4. An empty payload fingerprints to the empty string, not to `SHA-256("")`.**
This distinguishes "no payload at all" from "a payload whose hash happens to be
the empty-input digest", and avoids presenting a real-looking hash for missing
data. *Alternative considered:* always hash, even zero bytes — rejected for the
collision-of-meaning above.

**5. Fields are exported; immutability is by convention at this stage.**
A record is a simple data value, so exported fields keep it ergonomic and match
the plan. There are no mutators, but the type is not clamp-down immutable in Go's
type system. Flagged as a trade-off below.

**6. Errors are plain `fmt.Errorf` values for now.**
No caller needs to branch on *which* validation failed yet. *Alternative
considered:* sentinel or typed errors up front — deferred until `SyncSource`
(a later phase) demonstrates a need to distinguish failure kinds.

## Risks / Trade-offs

- **Exported mutable payload** → `RawData []byte` is stored as given, so a caller
  holding the slice can mutate it after construction and silently desync the
  `Fingerprint`. *Mitigation:* document the expectation that the payload is
  treated as immutable and not retained by callers; if it bites, copy the bytes
  on construction behind the existing constructor without changing callers.
- **Fingerprint stability is the source's responsibility** → if a source emits
  payloads with incidental ordering/nonce noise, fingerprints will flap and every
  sync will look like a change. *Mitigation:* normalize at the source adapter
  (Phase 7), before the record is constructed; the domain stays format-agnostic.
- **Convention-only immutability** → nothing at compile time stops future code
  from adding a mutator. *Mitigation:* keep mutation out of the package and treat
  "records are immutable" as a review rule; revisit if violations appear.

## Migration Plan

None required. The change is purely additive: a new type in an empty package. No
existing callers, no data, no rollout or rollback steps.
