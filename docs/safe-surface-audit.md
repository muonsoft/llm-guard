# Safe surface audit (M7)

Checklist for production-safe defaults on errors, observer events, formatting, and
JSON surfaces. Verified by marker-based tests in `audit_test.go`, `observer_test.go`,
and existing redaction tests.

## Safe-by-default surfaces

| Surface | Sensitive content excluded | Tests |
|---------|---------------------------|-------|
| `Event` (`%v`, `%+v`, `%#v`, JSON) | input text, findings, tokens, detector causes | `TestObserver_WhenEventFormatted_*`, `TestRedact_*` |
| `MaskResult` formatting/JSON | masked text values, finding values, placeholders | `TestRedact_WhenMaskResultAndTokenSetFormatted_*` |
| Safe `Event.EntityCounts` | caller-controlled custom entity strings | `TestObserver_WhenCustomEntityMarker_*` (bucketed as `CUSTOM`) |
| `TokenSet` formatting/JSON | token strings, mappings | `mask_test.go`, `TestRedact_*` |
| Public errors (`DetectorError`, `BlockError`, `InvalidTokenSetError`, …) | input text, matched substrings, detector causes | `TestRedact_WhenSafeErrorsFormatted_*`, existing `*_test.go` error suites |
| Default observer | no callbacks (`NoopObserver`) | `TestObserver_WhenNoObserverConfigured_*` |

## Lifecycle outcomes covered

- Detect: success, invalid UTF-8, detector error, cancellation, no detectors
- Mask: success, block, nested Detect event, entropy error paths (existing tests)
- Restore miss scanning uses the original response before replacement; repeated unknown placeholders count separately.

## Unsafe development path

| Requirement | Status |
|-------------|--------|
| Separate `Unsafe*` types and `WithUnsafeDevelopmentObserver` | implemented |
| Prominent `UNSAFE FOR PRODUCTION` GoDoc | `unsafe_observer.go` |
| No activation from `WithObserver` alone | `TestAudit_WhenSafeObserverOnly_*` |
| TokenSet mappings never exported via unsafe event | by design |

## Dependencies audit

`go.mod` contains only `github.com/muonsoft/errors`, `github.com/muonsoft/go-razdel`, and test `testify`. No Prometheus, OpenTelemetry, logging framework, exporter, network client, storage, or proxy dependency was added in M7.

## Residual limitations

- Unsafe diagnostics can leak sensitive text when explicitly enabled; callers must keep them out of production builds.
- Evaluation corpus is representative, not exhaustive; see family corpora for deep regression coverage.
- Benchmark numbers are not SLOs.
