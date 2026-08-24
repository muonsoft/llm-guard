# Evaluation data sources

How external evaluation data is pinned, cached, normalized, and attributed.
Evaluation commands read **local verified artifacts only**; network access is
confined to explicit `llmguard-eval-data -action fetch`.

## Layout

```text
testdata/evaluation/
  manifests/*.json          # source pins, digest, license, attribution
  mappings/*.json           # source label → product entity policy
  generated/*.jsonl         # fixed-seed offline suites (committed)
  thresholds/*.json         # profile gate/diagnostic policy
  cases.jsonl               # schema v1 conformance smoke

.cache/llm-guard/evaluation/   # gitignored raw + normalized data
  raw/<source-id>/...
  normalized/<source-id>.jsonl
```

Default cache path: `./.cache/llm-guard/evaluation` (override with `-cache`).

## Committed manifests

| ID | URL | License | Distribution |
| --- | --- | --- | --- |
| `redmadrobot-pii-benchmark` | [Hugging Face pii_benchmark](https://huggingface.co/datasets/redmadrobot-rnd/pii_benchmark) | MIT | cache-only |
| `factrueval-2016` | [factRuEval-2016](https://github.com/dialogue-evaluation/factRuEval-2016) | MIT | cache-only |
| `gitleaks-reviewed-fixtures` | [gitleaks](https://github.com/gitleaks/gitleaks) | MIT | committed-derived |

Attribution text is embedded in each manifest `attribution` field. Review
`license` and `distribution` before adding or refreshing a source.

ФИАС/ГАР address generation remains disabled pending separate license review.
FactRuEval uses clone-into-cache during fetch (optional; not required for the
committed RedMadRobot baseline).

## Fetch and normalize

```bash
# 1. Download and verify digest (network)
go run ./cmd/llmguard-eval-data -action fetch \
  -manifest ./testdata/evaluation/manifests/<source>.json \
  -cache ./.cache/llm-guard/evaluation

# 2. Normalize with mapping policy (network-free; re-verifies digest)
go run ./cmd/llmguard-eval-data -action normalize \
  -manifest ./testdata/evaluation/manifests/<source>.json \
  -mapping ./testdata/evaluation/mappings/<mapping>.json \
  -cache ./.cache/llm-guard/evaluation
```

RedMadRobot example:

```bash
go run ./cmd/llmguard-eval-data -action fetch \
  -manifest ./testdata/evaluation/manifests/redmadrobot-pii-benchmark-v1.json \
  -cache ./.cache/llm-guard/evaluation

go run ./cmd/llmguard-eval-data -action normalize \
  -manifest ./testdata/evaluation/manifests/redmadrobot-pii-benchmark-v1.json \
  -mapping ./testdata/evaluation/mappings/redmadrobot-v1.json \
  -cache ./.cache/llm-guard/evaluation
```

Output: `.cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark.jsonl`

Evaluate (no network):

```bash
go run ./cmd/llmguard-eval \
  -suite ./.cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark.jsonl \
  -profile contract -format json
```

## Adding a new source

1. **License review** — confirm redistribution mode (`cache-only`,
   `manifest-only`, or `committed-derived`).
2. **Pin revision** — record immutable `revision` and `digest_sha256` in a new
   manifest under `testdata/evaluation/manifests/`.
3. **Mapping policy** — versioned JSON under `testdata/evaluation/mappings/`
   with disposition rules (`supported`, `unsupported`, `ignored`).
4. **Adapter** — implement normalization in `internal/evaluation` (not public API).
5. **Golden tests** — adapter unit tests on Unicode, punctuation, and edge cases.
6. **Baseline** — after maintainer review, publish safe aggregates in
   `docs/evaluation/` (no raw text or failure-row dumps).
7. **Thresholds** — update or add threshold set; external profiles default to
   `diagnostic` until explicitly gated.

Do not vendor raw corpora into the Go module without a separate redistribution
review. Do not commit `.cache/` contents.

## CI lanes

| Lane | Trigger | Network | Purpose |
| --- | --- | --- | --- |
| PR `evaluation` + `evaluation-smoke` | push/PR | no | v1 corpus + generated smoke gates |
| `evaluation-external.yml` | weekly / manual | yes (fetch) | RedMadRobot exposure diagnostic artifact |

External workflow uploads JSON artifacts only; it does **not** auto-commit
baselines.

## Related docs

- [suite-v2.md](suite-v2.md) — schema v2 fields and metric formulas
- [external-baseline.md](external-baseline.md) — committed RedMadRobot aggregates
- [../evaluation-baseline.md](../evaluation-baseline.md) — schema v1 smoke
