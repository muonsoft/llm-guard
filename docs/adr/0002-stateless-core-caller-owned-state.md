# ADR 0002: Stateless core with caller-owned future state

## Status

Accepted (M0)

## Context

LLM Guard will eventually support reversible masking, token sets, and session-scoped
restore maps. Those features require caller-controlled persistence and lifecycle.
The M0 milestone delivers detection-only functionality that must remain embeddable
without imposing storage, logging, or provider dependencies.

## Decision

The root `llmguard` package is **stateless after construction**:

- `Guard` is immutable once `New` returns; detectors are fixed at build time.
- `Detect` performs validation and aggregation only; it does not retain input
  text, findings, or error causes in package-level state.
- Masking placeholders, restore maps, audit trails, and metrics are **out of scope**
  for M0 and will live in caller code or future packages built on top of the
  detection contract.

Runtime registration of detectors is rejected to avoid locks and temporal coupling
between concurrent `Detect` calls.

## Consequences

- Callers own any state needed for masking and restore in later milestones.
- The library remains safe for concurrent use without hidden global mutation.
- Future masking layers can consume stable, validated findings without changing
  M0 detection semantics.
