# ADR 0001: UTF-8 byte offsets for finding spans

## Status

Accepted (M0)

## Context

Findings must reference spans in the original input text without storing the
matched value. Go string indexing, regular expressions, and substring operations
all work in UTF-8 byte offsets. Downstream masking will slice `text[start:end]`
directly.

## Decision

`Finding.Start` and `Finding.End` use **UTF-8 byte offsets** forming a half-open
interval `[Start, End)`. The core validates that both boundaries align with UTF-8
rune boundaries. `End == len(text)` is allowed.

Rune offsets are not used in the public API for M0. Helpers may be added later if
callers need rune-based reporting.

## Consequences

- Detectors can pass through byte indexes from Go regexp without conversion.
- Masking and restoration logic can slice strings directly.
- Invalid spans that split a multibyte rune are rejected before aggregation.
- API documentation and examples must state byte-offset semantics explicitly.
