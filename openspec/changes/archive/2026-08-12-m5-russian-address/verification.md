# Verification Report: m5-russian-address

## Summary

| Dimension | Status |
|---|---|
| Completeness | 33/33 tasks complete; 8/8 requirements implemented; canonical specs synced and strictly validated |
| Correctness | 20/20 scenarios covered by implementation, focused tests, corpus or dependency evidence |
| Coherence | Design decisions followed; no unresolved contradiction or project-pattern deviation |

## Completeness

### Task coverage

- Core runtime, detector, resolution/masking, corpus/audit, tests and docs tasks are complete.
- Strict change validation passed after implementation.
- Task 7.2 passed after the delta specs were synced to the canonical specs.
- Task 7.3 is satisfied by this independently reviewed verification report.

### Requirement coverage

1. **Immutable RU ADDRESS detector** — `address.go` implements the stable opt-in `NewAddressDetector`; cancellation and concurrency coverage is in `address_test.go` and `internal/nlp/address_context_test.go`.
2. **Supported street and house compositions** — `internal/nlp/address.go` implements bounded prefix/suffix labels, street names, labeled/unlabeled houses, settlement prefix and maximal corpus/building/apartment suffixes. `address_test.go`, `internal/nlp/address_test.go` and the corpus cover direct, compact, reordered and bounded identifier forms.
3. **Conservative acceptance** — street+house is mandatory. Standalone parts, settlement+street, embedded tokens, malformed separators and invalid identifiers are negative cases in `address_test.go` and `testdata/address/cases.jsonl`.
4. **Exact UTF-8 spans** — composition uses tokenizer source byte offsets and rune-safe outer boundaries; Unicode punctuation, sentence boundaries and multiple occurrences have exact-slice tests.
5. **Pure-Go runtime boundary** — production uses the existing pinned `go-razdel` adapter and project-authored finite rules. `go list -deps`, `go mod graph` and `go mod tidy -diff` show no new dependency or runtime data.
6. **Versioned quality boundary** — `testdata/address/cases.jsonl` contains 33 synthetic positive/negative/ambiguous cases; `address_corpus_test.go` reports TP=20, FP=0, FN=0, precision/recall/exact-span-rate 1.000 and checks all R0 ADDRESS links against pinned fixtures.
7. **Mask/Restore pipeline** — `address_test.go` verifies one token over the full accepted span, byte-for-byte restoration and distinct caller-owned TokenSets under concurrency.
8. **ADDRESS over nested PERSON** — the resolver's central priority table already ranks ADDRESS above PERSON; permutation and Guard end-to-end tests preserve the full `ул. Академика Сахарова, 10` finding without PERSON leakage.

## Correctness

All 20 scenarios map to direct evidence:

- registration, pre/post-tokenization and traversal cancellation, shared-detector concurrency;
- mandatory, extended, compact, suffix-label and bounded house-identifier forms;
- standalone/ambiguous/embedded/malformed negatives;
- Unicode byte slices and stable multiple-address order;
- offline Go build/dependency graph and pinned Natasha fixture verification;
- deterministic corpus evaluation and unexplained-differential rejection;
- literal Mask/Restore round trip and independent concurrent TokenSets;
- ADDRESS/PERSON priority under candidate permutations and the real Guard pipeline.

Independent verification commands:

- `go test ./... -run 'Test(Address|Addr|Resolve.*Address|Mixed|ValidHouseIdentifier)' -count=1` — PASS.
- `go test -race ./... -run 'TestAddress' -count=1` — PASS.
- `go test ./... -run 'TestAddressCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors' -v -count=1` — PASS; TP=20, FP=0, FN=0, no unexplained R0 differential.
- `PYTHONDONTWRITEBYTECODE=1 python3 tools/natasha-reference/reference.py verify --offline ...` — PASS.
- `go test ./... -count=1 -timeout 120s` — PASS.
- `go vet ./...` — PASS.
- `go test -race ./... -count=1` — PASS.
- `go list -deps ./...`, `go mod graph`, `go mod tidy -diff` — PASS; no new production dependency.
- `openspec validate m5-russian-address --strict --no-interactive` — PASS.
- `git diff --check` — PASS.

## Coherence

- The implementation extends the existing internal tokenizer runtime rather than introducing regex-only whole-address parsing or a new grammar dependency.
- The explicit street+house matrix remains the only acceptance path; optional parts only extend an already accepted composition.
- Internal candidate parts never escape as public entities; the public API returns only full ADDRESS findings.
- ADDRESS/PERSON conflict handling stays in the common resolver rather than coupling detectors.
- Corpus and quality documentation remain synthetic and development-only; original address values are not stored in global state or error output.
- The implementation remains opt-in and additive, preserving existing public interfaces and default Guard configuration.

## Issues by priority

### CRITICAL

- None.

### WARNING

- None.

### SUGGESTION

- None.

## Orchestration evidence

- Variant C resumed the original blocked milestone without switching transport or duplicating the OpenSpec change.
- One Composer 2.5 primary job was followed by one consolidated review correction and one narrow verification-evidence correction.
- Full review found compound identifier, suffix-label, separator, traversal-cancellation, TokenSet-evidence and repository-hygiene gaps; the verification correction then bounded newly enabled multi-segment identifiers.
- Final Herdr `agent get` and bounded `agent read` completed; one stalled Cursor UI operation was cancelled with `Esc` and the same session resumed successfully.

## Final assessment

Implementation, all requirements/scenarios, design coherence, ADDRESS/PERSON overlap policy, corpus quality and broad verification are green. No CRITICAL or WARNING remains. Ready for spec sync and post-sync validation.
