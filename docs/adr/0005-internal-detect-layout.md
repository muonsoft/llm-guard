# ADR 0005: Built-in detectors live under internal/detect

## Status

Accepted

## Context

After MVP, the root `llmguard` package mixed the stable public API from
[ADR 0004](0004-mvp-public-api-boundary.md) with private detector bodies and
shared scanning helpers. PERSON and ADDRESS already called `internal/nlp` and
mapped spans in the root package. Structured PII and secrets did not.

The original MVP plan sketched public `detector/`, `resolver/`, and `masking/`
packages, then asked to keep the public API compact.

## Decision

- Keep the canonical import path `github.com/muonsoft/llm-guard` as the only
  supported consumer surface.
- Implement built-in structured PII and secret scanners in `internal/detect`.
  That package returns `Span` values and must not import `llmguard`.
- Keep public constructors (`NewEmailDetector` and the rest) in the root
  package. They map spans to `Finding`, including entity, detector name, and
  confidence.
- Reuse the same root adapter for PERSON and ADDRESS over `internal/nlp`.
- Do not introduce public subpackages for detectors, resolver, or masking.
- Do not alias public types onto `internal/` types.

## Rejected alternatives

- **Public `detect/` packages:** expands the compatibility surface without a
  second consumer for those APIs.
- **Internal types that implement `llmguard.Detector`:** import cycle with root
  constructors.
- **Type aliases to `internal/core`:** worse GoDoc and unnecessary risk to the
  ADR 0004 contracts.

## Consequences

- External consumers continue to depend only on the root module path.
- Detector internals can change without a public API revision.
- Contract tests and fuzz targets remain in the root package so CI target names
  stay stable.
