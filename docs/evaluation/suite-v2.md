# Suite schema v2

Normalized evaluation suites (schema version 2) support heterogeneous external
annotations, source-to-product mapping, and separate contract, exposure, and
lifecycle profiles. They complement — but do not replace — the strict schema v1
corpus at `testdata/evaluation/cases.jsonl`.

## Record fields

Each JSONL line is one `SuiteRecord`:

| Field | Required | Description |
| --- | --- | --- |
| `schema_version` | yes | Must be `2` |
| `suite_id` | yes | Logical suite identifier |
| `source_id` | yes | Source manifest ID |
| `source_record_id` | yes | Stable record key within the source |
| `mapping_version` | yes | Mapping policy version used during normalization |
| `input` | yes | UTF-8 input text |
| `input_sha256` | yes | SHA-256 of `input` bytes (integrity check) |
| `annotations` | yes | Array of source spans (may be empty or `null`) |
| `declared_entities` | no | Optional entity scope hint |
| `lifecycle` | no | Optional mask/restore expectations |

### Annotation fields

| Field | Description |
| --- | --- |
| `source_label` | Label from the external source |
| `mapped_entity` | Product entity when disposition is `supported` |
| `start`, `end` | Half-open UTF-8 byte interval `[start, end)` |
| `disposition` | `supported`, `unsupported`, or `ignored` |
| `reason` | Required when `disposition` is `ignored` |

### Dispositions

- **`supported`** — participates in contract (exact span match) and exposure
  (byte union). Requires `mapped_entity`.
- **`unsupported`** — exposure only; does not affect contract TP/FN.
- **`ignored`** — excluded from scoring (adapter/annotation defect). Requires
  `reason`. Published separately; must not be used to inflate scores.

## Profiles

Run with `go run ./cmd/llmguard-eval -suite <path> -profile <name>`.

| Profile | Measures | Typical gate |
| --- | --- | --- |
| `contract` | Exact `(entity, start, end)` TP/FP/FN on `supported` annotations | Offline smoke / v1 corpus |
| `exposure` | Byte-interval union coverage of sensitive surface | Diagnostic on external holdout |
| `lifecycle` | Mask → LLM response → Restore under default policy | Offline generated smoke |

Optional `-thresholds` loads a versioned threshold set (see
`testdata/evaluation/thresholds/`).

## Contract formulas

On `supported` annotations within per-record entity scope:

| Metric | Formula | Zero denominator |
| --- | --- | --- |
| Precision | TP / (TP + FP) | 0 |
| Recall | TP / (TP + FN) | 0 |
| F1 | 2PR / (P + R) | 0 |

Match key: exact entity and UTF-8 byte span after `Detect → Resolve`.

`-fail-on-regression` exits non-zero when any in-scope FP or FN occurs (no
threshold file required).

## Exposure formulas

Gold sensitive intervals `G` are built from `supported` and `unsupported`
annotations (excluding `ignored`). Predicted intervals `P` come from resolved
findings (any entity covers bytes).

```text
sensitive_bytes        = |G|
covered_sensitive      = |G ∩ P|
leaked_sensitive       = |G \ P|
overmatched_bytes      = |P \ G|
byte_coverage          = covered_sensitive / sensitive_bytes   (0 if |G| = 0)
```

Overlapping source annotations and predictions are merged at the byte level.
Prediction entity need not match source label for byte coverage: masking bytes
under a different entity still counts as covered.

## Lifecycle expectations

When `lifecycle` is present:

- `expected_action` — `mask` or `block`
- `response_recipe` — how the simulated LLM response is built (e.g. `identity`,
  `mutate_placeholder`)

Failures emit safe diagnostics (record ID, spans, labels, hashes) — never raw
`input` substrings in committed reports.

## Safe diagnostics policy

External and lifecycle formatters MUST NOT include:

- raw `input` text
- gold or predicted substrings
- arrays of per-case failure rows in committed baseline artifacts

Committed baselines publish aggregates, counts, and metadata only. Full
diagnostics may appear in local/CI artifacts but are not vendored into git.

## Related docs

- [sources.md](sources.md) — fetch/normalize pipeline and manifests
- [external-baseline.md](external-baseline.md) — RedMadRobot diagnostic baseline
- [../evaluation-baseline.md](../evaluation-baseline.md) — schema v1 smoke corpus
