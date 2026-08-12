## Verification Report: m3-structured-completeness

### Summary

| Dimension | Status |
|---|---|
| Completeness | 30/30 tasks evidenced; 9/9 delta requirements and 22/22 scenarios implemented |
| Correctness | Focused, two exact fuzz targets, full tests, vet and race checks pass independently |
| Coherence | Implementation follows context-first design, stable API ADR and existing reversible pipeline |

### Completeness

- PASSPORT, BANK_ACCOUNT and DATE_OF_BIRTH public constructors and immutable detectors exist in dedicated root-package files.
- `contextmatch.go` centralizes bounded same-line RU marker handling without storing input or normalized values.
- `RegexDetectorConfig`, constructor-only `CustomRegexpDetector`, safe configuration errors and core-filled detector metadata implement the new custom-detection capability.
- Resolver fallback remains internal and deterministic; built-in/custom overlap regressions pass without public priority configuration.
- Mixed round trip, per-entity corpus, cancellation, concurrency, boundary, error-safety and public example coverage are present.
- ADR 0004 records the accepted M0–M3 public API boundary and additive extension rules for M4–M7.

### Correctness mapping

| Requirement group | Implementation evidence | Scenario evidence |
|---|---|---|
| Remaining structured public pack | `passport.go`, `bankaccount.go`, `dateofbirth.go` | stable-name, cancellation and concurrency tests in each detector test file; mixed round trip in `structured_test.go` |
| PASSPORT context/forms | `passport.go`, `contextmatch.go` | `passport_test.go`; positive/negative/malformed corpus rows in `structured_test.go` |
| BANK_ACCOUNT fallback/BIK checksum | `bankaccount.go`, `contextmatch.go` | before/after BIK, invalid checksum, false label, unrelated nine-digit and grouped tests in `bankaccount_test.go` |
| DATE_OF_BIRTH context/calendar/boundaries | `dateofbirth.go`, `contextmatch.go` | numeric/textual, ordinary, invalid, embedded and multiline tests in `dateofbirth_test.go` |
| Structured corpus completion | `structured_test.go`, `fuzz_test.go` | per-entity counts plus `FuzzStructuredDetectorsInvariants` |
| Regexp configuration and safe errors | `regexp_detector.go`, `errors.go`, `validate.go` | empty/malformed, NaN/Inf and no-value-leak tests in `regexp_detector_test.go` |
| Regexp matching/zero-width | `regexp_detector.go` | direct UTF-8 spans, caller boundaries, zero-width unit and fuzz coverage |
| Custom reversible integration | unchanged Guard/Resolve/Mask/Restore pipeline | regexp and external Go detector round trips plus built-in overlap tests |
| Custom concurrency/error behavior | `regexp_detector.go` and existing core validation | cancellation, concurrency and unsafe external finding tests |

### Coherence

- All new structured detectors use candidate → normalize/validate → Finding and return only exact source spans.
- BANK_ACCOUNT performs checksum only for a boundary-valid explicitly labeled same-line BIK; otherwise the documented marker-required fallback applies.
- DATE_OF_BIRTH is not a generic date detector and rejects ordinary/contract, embedded, impossible and multiline forms.
- Custom regexp uses Go RE2, compiles once, adds no implicit boundaries and leaves metadata normalization to Guard.
- No external service, new dependency, mutable registry, provider adapter, persistence, logging or secret value enters the root core.
- Public API changes are additive and conform to `docs/adr/0004-mvp-public-api-boundary.md`.

### Issues by priority

#### CRITICAL

- None.

#### WARNING

- None.

#### SUGGESTION

- None required for M3 acceptance.

### Verification commands

- `go test ./... -run 'Test(Passport|BankAccount|DateOfBirth|Regex|Custom)'` — PASS.
- `go test ./... -run 'Test(MixedStructured|StructuredCorpus|Structured_WhenConcurrent)'` — PASS.
- `go test . -run '^$' -fuzz '^FuzzStructuredDetectorsInvariants$' -fuzztime=2s` — PASS (7,493 executions; 6 new interesting inputs in independent run).
- `go test . -run '^$' -fuzz '^FuzzCustomRegexpDetectorInvariants$' -fuzztime=2s` — PASS (26,743 executions; 4 new interesting inputs in independent run).
- `go test ./...` — PASS.
- `go vet ./...` — PASS.
- `go test -race ./...` — PASS.
- `git diff --check` — PASS.
- `openspec validate m3-structured-completeness --strict --no-interactive` — PASS before sync.

### Orchestration evidence

- Variant C primary: one Composer 2.5 job.
- Corrections: one consolidated correction plus one verification-evidence correction from newly revealed BANK_ACCOUNT test evidence.
- Herdr transport: degraded after one bounded post-correction read timeout; recovered on the single liveness retry; all later get/read calls succeeded; no ambiguous completion.
- Primary and correction result files are under `.agent-orchestration/results/` and are intentionally not versioned.

### Final assessment

Implementation, scenario coverage and design coherence are clean with no CRITICAL or WARNING findings. Ready for OpenSpec verify, spec sync and archive.
