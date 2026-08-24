# External prefilter baseline (RedMadRobot PII benchmark)

Safe aggregate evidence for llm-guard as a **precision-oriented prefilter** on an
external Russian PII holdout. This report documents profile scope, reproduction
commands, and measured aggregates only. It does **not** claim completeness,
high-recall DLP coverage, or a release SLO on external F1.

Machine-readable metadata: [external-baseline.json](external-baseline.json).

## Positioning

llm-guard is a local **precision-oriented prefilter**: it reduces the risk of
sending documented supported PII and secret forms to an LLM through reversible
masking. It does **not** replace high-recall DLP, generic NER, or domain-specific
security review. External contract metrics on a wider annotation scope are
**diagnostic**; offline gates remain schema v1 conformance and committed
generated/lifecycle smoke.

## Measurement context

| Field | Value |
| --- | --- |
| Measured commit | `c351ecc65412313de67db3ca84c9a0eb8945e257` |
| Go / OS | `go1.26.2`, `linux/amd64` |
| Source | RedMadRobot `pii_benchmark` (`redmadrobot-pii-benchmark`) |
| Manifest revision | `f77ea831274daf980cc45c61a93c226be9d978d6` |
| Raw `test.csv` SHA-256 | `6bf544a380a3ee5bec94b946124bea3afaecce49e734679ad0f0c0e7c12977bb` |
| Mapping version | `redmadrobot-v1` |
| Normalized records | 2841 |
| Ignored records | 2 (`adapter_token_not_found_in_text`) |

## Reproduction commands

Data preparation (network required; cache is gitignored):

```bash
go run ./cmd/llmguard-eval-data -action fetch \
  -manifest ./testdata/evaluation/manifests/redmadrobot-pii-benchmark-v1.json \
  -cache ./.cache/llm-guard/evaluation

go run ./cmd/llmguard-eval-data -action normalize \
  -manifest ./testdata/evaluation/manifests/redmadrobot-pii-benchmark-v1.json \
  -mapping ./testdata/evaluation/mappings/redmadrobot-v1.json \
  -cache ./.cache/llm-guard/evaluation
# → .cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark.jsonl
```

Evaluation (network-free; reads normalized cache only):

```bash
go run ./cmd/llmguard-eval \
  -suite ./.cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark.jsonl \
  -profile exposure -format json

go run ./cmd/llmguard-eval \
  -suite ./.cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark.jsonl \
  -profile contract -format json
```

See [sources.md](sources.md) for manifest, license, and cache layout.

## Exposure profile (diagnostic)

Profile: **exposure** on supported + unsupported source annotations (byte-interval
union). Status: **diagnostic** — not a release gate.

| Metric | Value |
| --- | ---: |
| sensitive_bytes | 87279 |
| covered_sensitive_bytes | 14542 |
| leaked_sensitive_bytes | 72737 |
| overmatched_bytes | 678 |
| byte_coverage | 0.167 |

## Contract profile (diagnostic)

Profile: **contract** on `supported` annotations only (exact entity/span match).
Status: **fail** on this holdout — **expected** and **not** a release SLO.
Aggregate: TP=662, FP=119, FN=1717.

Per-entity aggregates (do not treat as DLP guarantee):

| Entity | TP | FP | FN | Precision | Recall |
| --- | ---: | ---: | ---: | ---: | ---: |
| EMAIL | 62 | 0 | 159 | 1.00 | 0.28 |
| PHONE | 103 | 2 | 68 | 0.98 | 0.60 |
| INN | 192 | 4 | 71 | 0.98 | 0.73 |
| IP_ADDRESS | 94 | 4 | 61 | 0.96 | 0.61 |
| URL | 116 | 2 | 88 | 0.98 | 0.57 |
| BANK_CARD | 61 | 23 | 142 | 0.73 | 0.30 |
| ADDRESS | 27 | 37 | 113 | 0.42 | 0.19 |
| PERSON | 2 | 32 | 303 | 0.06 | 0.007 |
| PASSPORT | 4 | 15 | 490 | 0.21 | 0.008 |
| SNILS | 1 | 0 | 222 | 1.00 | 0.004 |

## Threshold policy (`prefilter-v1`)

Threshold set: `testdata/evaluation/thresholds/prefilter-v1.json`.

| Profile | Status | Rationale |
| --- | --- | --- |
| contract (external) | diagnostic | Unknown empirical holdout must not become an SLO |
| exposure (external) | diagnostic | Byte coverage is evidence, not a DLP guarantee |
| lifecycle (generated smoke) | gate | Zero unexpected failures on committed offline smoke |
| schema v1 corpus | gate (existing) | Strict conformance on documented MVP contract |

Offline PR/release gates: `go test ./...`, v1 `-corpus -fail-on-regression`, and
generated contract/lifecycle smoke with `-fail-on-regression`. External lanes run
on schedule or `workflow_dispatch` only.

## Limitations

- **Unsupported labels** — OMS, driver license, military ID, birth certificate,
  and other absent product entities remain exposure-only.
- **Conservative PERSON** — single given names or surnames without supported
  multi-token FIO patterns are largely missed by design.
- **Conservative ADDRESS** — city-only, region-only, or street-without-house
  spans are not contract obligations.
- **Checksum-invalid PII** — invalid INN/SNILS checksums are intentionally
  excluded from supported contract scope.
- **Not DLP** — low PERSON/PASSPORT/SNILS recall on this holdout reflects
  intentional prefilter boundaries, not a promise to detect all PII.

For operational guidance on high-risk deployments, see
[known-limitations.md](../known-limitations.md).
