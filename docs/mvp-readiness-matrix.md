# MVP Definition of Done matrix (M8)

Evidence map for `v0.1.0` readiness. Status reflects repository state after M8
OSS stabilization. Tag/release publication is tracked separately in
[release-checklist.md](release-checklist.md).

| # | Criterion | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Pure Go library | met | `go.mod`; no cgo in module |
| 2 | No mandatory external service | met | local detectors only; [README](../README.md) |
| 3 | Core API independent of OpenAI | met | no provider packages in module |
| 4 | Structured PII detectors work | met | `go test ./...`; family corpora under `testdata/` |
| 5 | PERSON rule-based conservative FIO | met | [docs/person-quality-report.md](person-quality-report.md), `testdata/person/cases.jsonl` |
| 6 | ADDRESS compositional only | met | [docs/address-quality-report.md](address-quality-report.md), `testdata/address/cases.jsonl` |
| 7 | `ORGANIZATION` absent | met | no detector symbol in public API |
| 8 | Custom regex entities | met | `NewCustomRegexpDetector`, `example_test.go` |
| 9 | User-defined `Detector` interface | met | `Detector` in `guard.go` |
| 10 | Basic secret detection | met | JWT/PEM/API key/DSN detectors; [secret-patterns.md](secret-patterns.md) |
| 11 | Deterministic resolver | met | `resolve.go`, `FuzzResolveInvariants` |
| 12 | Reversible masking | met | `mask.go`, `FuzzMaskRestoreRoundTrip` |
| 13 | Caller-owned RAM `TokenSet` | met | `doc.go`, `placeholder.go`, `audit_test.go` |
| 14 | Safe audit mode default | met | `NoopObserver` default; [safe-surface-audit.md](safe-surface-audit.md) |
| 15 | PII not in standard logs/errors | met | `audit_test.go`, observer tests |
| 16 | `go test -race ./...` passes | met | `./scripts/release-check.sh` race gate; CI `race` job |
| 17 | Fuzz tests mask/restore/resolver | met | `fuzz_test.go`; `./scripts/release-check.sh fuzz` |
| 18 | Evaluation corpus | met | `testdata/evaluation/cases.jsonl`, `cmd/llmguard-eval` |
| 19 | Per-entity metrics in evaluation | met | evaluation report entity table |
| 20 | Natasha subset documented | met | [natasha-license-inventory.md](natasha-license-inventory.md) |
| 21 | Python Natasha reference-only | met | inventory decision rows; no embedded dicts |
| 22 | Dependency/dictionary licenses verified | met | [dependency-license-inventory.md](dependency-license-inventory.md), `./scripts/release-check.sh license` |
| 23 | README embedded Go example | met | [README](../README.md), `example_test.go` |
| 24 | External module consumer check | met | `testdata/external-consumer/`, `./scripts/release-check.sh consumer` |
| 25 | OSS policies published | met | [SECURITY.md](../SECURITY.md), [CONTRIBUTING.md](../CONTRIBUTING.md) |
| 26 | Release dry-run without publication | met | `scripts/release-check.sh`, `.github/workflows/release-check.yml` |
| 27 | Known limitations published | met | [known-limitations.md](known-limitations.md) |
| 28 | Quality/benchmark regression vs M7 | met | [m8-quality-benchmark-comparison.md](m8-quality-benchmark-comparison.md) |

## Release boundary

- Repository readiness for `v0.1.0`: **yes** when `./scripts/release-check.sh` is green.
- Published `v0.1.0` tag/release exists: **no** until manual steps in
  [release-checklist.md](release-checklist.md).

## Out of scope (confirmed absent)

- OpenAI adapters / HTTP proxy
- Persistent audit store / exporter server
- ML/NER beyond rule-based MVP
- Automatic tag/push/release from CI
