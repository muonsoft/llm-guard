# ADR 0004: Stable MVP public API boundary after structured completeness

## Status

Accepted (M3)

## Context

M0–M3 established the embeddable core, deterministic resolution, reversible
masking, the complete structured PII pack, custom Go detectors and a safe custom
regexp adapter. M4–M7 add Russian NLP detectors, secrets, policy and observability.
Before those layers land, the existing public surface needs an explicit
compatibility boundary so later milestones do not reshape the core implicitly.

The review covered every exported symbol in the root `llmguard` package, the
canonical OpenSpec requirements, examples, error behavior, concurrency contracts
and caller-owned state.

## Decision

The following contracts are stable for the remainder of the MVP:

- `EntityType` remains an extensible string type. New built-in constants are
  additive; callers may use their own non-empty values.
- `Finding` keeps UTF-8 byte offsets in the half-open interval `[Start, End)`, a
  finite confidence in `[0,1]`, entity and detector metadata, and never stores
  the matched value.
- `Detector` remains the two-method extension point. Its name is stable,
  non-sensitive and non-empty; `Detect` accepts `context.Context`, returns source
  spans, and implementations registered in a Guard are concurrent-safe.
- `New` plus immutable `Option` values construct a concurrent-safe `Guard`.
  Runtime detector registration and package-global mutable configuration are not
  added during the MVP.
- `Detect`, public `Resolve`, `Mask` and `Restore` retain their current pipeline
  roles. Resolver priority is internal but its deterministic ordering and
  built-in-over-custom overlap behavior are observable contracts.
- `TokenSet` remains opaque and caller-owned. It does not expose mappings or raw
  values and keeps redacted formatting/JSON behavior.
- Exported sentinel and typed errors remain usable through
  `github.com/muonsoft/errors` `Is`/`As`. Human-readable error text is not a
  machine-readable compatibility surface and must remain free of sensitive input.
- Built-in detector constructors may be added without changing `Detector`.
  `CustomRegexpDetector` is constructor-only, immutable after construction and
  returns full non-overlapping RE2 matches; its zero value is not usable.

M4–M7 may add entity constants, detector constructors, options and additive
types. They may extend the internal priority table only with explicit overlap
requirements and permutation tests. Policy and observability must consume the
validated/resolved pipeline rather than changing `Finding`, `Detector`, Guard
immutability, byte-offset semantics or TokenSet ownership.

Any proposal to change one of those stable contracts requires a new OpenSpec
change and ADR that names migration impact. In particular, changing spans to rune
indexes, closing `EntityType`, adding methods to `Detector`, mutating a live Guard,
exposing TokenSet mappings, returning partial findings on failure, or making
external services mandatory is breaking and cannot be folded into a later MVP
milestone as mechanical fallout.

## Rejected alternatives

- **Freeze every exported symbol exactly:** additive detector constructors,
  entities and options are necessary for the remaining MVP and do not break
  callers.
- **Expose a public priority registry now:** no accepted use case requires live
  priority mutation, while it would weaken deterministic concurrent behavior.
- **Add a generic plugin or policy interface before M6:** the current `Detector`
  extension point and pipeline are sufficient; premature interfaces would expand
  compatibility obligations without implementations.
- **Return only interfaces from every constructor:** built-in stateless detectors
  can remain hidden behind `Detector`, while the configured regexp adapter keeps
  a concrete exported type for discoverability and an explicit constructor-only
  contract.

## Consequences

- Existing M0–M3 callers remain source-compatible as new MVP detectors and layers
  are added.
- M4/M5 can implement PERSON and ADDRESS as ordinary immutable detectors on the
  established UTF-8 span contract.
- M6/M7 must add policy and safe observability around resolved findings without
  leaking values or introducing hidden core state.
- Observable resolver priority changes require specification and regression
  evidence even though the table itself remains internal.
- A future need for runtime registration, custom priority, persistent TokenSet or
  provider integration is a separate post-MVP design decision.
