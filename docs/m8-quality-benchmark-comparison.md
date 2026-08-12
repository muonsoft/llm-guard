# M8 quality and benchmark comparison

Final M8 regression comparison against the M7 release-candidate baselines.
This report documents reproduction commands and outcomes; it does **not** define
production SLOs or numeric pass/fail thresholds for benchmarks.

## Environment

| Field | M7 baseline | M8 verification |
| --- | --- | --- |
| Date | 2026-08-12 | 2026-08-12 |
| Go | go1.26.2 linux/amd64 | go1.26.2 linux/amd64 |
| OS | Linux 5.15.0-181-generic | Linux 5.15.0-181-generic |
| CPU | AMD EPYC-Rome Processor | AMD EPYC-Rome Processor |

Hardware variance applies; compare stable benchmark **names**, not absolute ns/op
across machines or dates.

## Quality regression

### Command

```bash
go run ./cmd/llmguard-eval \
  -corpus ./testdata/evaluation/cases.jsonl \
  -format json \
  -fail-on-regression
```

Also enforced in `./scripts/release-check.sh` (full mode) and CI `evaluation` job.

### M7 committed baseline

See [evaluation-baseline.md](evaluation-baseline.md):

- Cases: 34
- Aggregate: TP=22 FP=0 FN=0
- Coverage complete: true
- All 16 MVP entities reported

### M8 expectation

- Same corpus path and matching rules (exact UTF-8 byte spans after Detect → Resolve)
- Exit code zero with `-fail-on-regression`
- No new FP/FN or incomplete entity coverage

If M8 changes evaluation outcomes, update `docs/evaluation-baseline.md` only with
an explicit maintainer-approved baseline revision (not done in M8 implementation).

## Benchmark smoke

### Command

```bash
go test ./... -run '^$' -bench . -benchmem -count=1
```

Full dry-run uses `-count=1` for smoke; M7 baseline used `-count=5` for a richer
development snapshot ([benchmark-baseline.md](benchmark-baseline.md)).

### Stable benchmark names (unchanged M7 → M8)

| Group | Names |
| --- | --- |
| Detect | `BenchmarkDetect_RUPrompt`, `BenchmarkDetect_MixedPII`, `BenchmarkDetect_SyntheticSecret` |
| Mask | `BenchmarkMask_RUPrompt`, `BenchmarkMask_MixedPII`, `BenchmarkMask_SyntheticSecret` |
| Restore | `BenchmarkRestore_RUPrompt`, `BenchmarkRestore_MixedPII`, `BenchmarkRestore_SyntheticSecret` |
| Observer | `BenchmarkObserver_DefaultNoop`, `BenchmarkObserver_WithCallback` |

### M8 gate

- All listed benchmarks execute successfully in smoke mode
- No rename/removal without changelog and baseline update
- Numeric drift is informational only (no SLO)

## Fuzz smoke (M8 addition to release gate)

```bash
./scripts/release-check.sh fuzz
```

Targets (exact names, 2s each by default):

- `FuzzMaskRestoreRoundTrip`
- `FuzzResolveInvariants`
- `FuzzCustomRegexpDetectorInvariants`

## External consumer (M8 addition)

```bash
./scripts/release-check.sh consumer
```

Confirms documented Detect → Mask → Restore flow from an independent module using
only `github.com/muonsoft/llm-guard`.

## Summary

| Gate | M7 evidence | M8 release dry-run |
| --- | --- | --- |
| Evaluation regression | [evaluation-baseline.md](evaluation-baseline.md) | `release-check.sh` / CI |
| Benchmark names stable | [benchmark-baseline.md](benchmark-baseline.md) | benchmark smoke |
| Fuzz mask/restore/resolver | archive milestone verification | `release-check.sh fuzz` |
| External consumer | not required at M7 | `release-check.sh consumer` |
| License inventory | R0 + M7 dep audit note | `release-check.sh license` |

M8 adds distribution and supply-chain gates without changing runtime public API or
the M7 quality/benchmark semantics documented above.
