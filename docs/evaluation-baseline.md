# Evaluation baseline (M7 release candidate)

Representative full-MVP corpus at `testdata/evaluation/cases.jsonl`. This baseline
documents the unified runner boundary; deeper family corpora remain authoritative
for PERSON (`testdata/person/cases.jsonl`), ADDRESS (`testdata/address/cases.jsonl`),
and secrets (`testdata/secrets/cases.jsonl`).

## Reproduction command

```bash
go run ./cmd/llmguard-eval \
  -corpus ./testdata/evaluation/cases.jsonl \
  -format markdown \
  -fail-on-regression
```

JSON output:

```bash
go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format json
```

`-fail-on-regression` exits non-zero when aggregate FP or FN is greater than zero or corpus coverage is incomplete.

## Matching and formulas

Runner flow: **Detect → Resolve** on a Guard with all 16 built-in MVP detectors.
Match key: exact `(entity, start, end)` UTF-8 byte span.

| Metric | Formula | Zero denominator |
|--------|---------|------------------|
| Precision | TP / (TP + FP) | 0 |
| Recall | TP / (TP + FN) | 0 |
| F1 | 2PR / (P + R) | 0 |
| FPR | false_positive_cases / negative_cases | 0 |
| FNR | FN / (TP + FN) | 0 |

`negative_cases` counts corpus rows with no expected span for the entity.
`false_positive_cases` counts such rows where at least one predicted span exists.

## Committed report (2026-08-12)

Environment: `go1.26.2`, `linux/amd64`.

```
# llm-guard MVP evaluation report

Cases: 34

Matching uses Detect → Resolve with exact `(entity, start, end)` UTF-8 byte spans.

Formulas: precision = TP/(TP+FP), recall = TP/(TP+FN), F1 = 2PR/(P+R), FPR = false_positive_cases/negative_cases, FNR = FN/(TP+FN). Zero denominators yield 0.

| Entity | TP | FP | FN | Neg cases | FP cases | Precision | Recall | F1 | FPR | FNR |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| PERSON | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| ADDRESS | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| EMAIL | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| PHONE | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| IP_ADDRESS | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| URL | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| INN | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| SNILS | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| PASSPORT | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| BANK_CARD | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| BANK_ACCOUNT | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| DATE_OF_BIRTH | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| SECRET_JWT | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| SECRET_PRIVATE_KEY | 1 | 0 | 0 | 33 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| SECRET_API_KEY | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |
| CONNECTION_STRING | 2 | 0 | 0 | 32 | 0 | 1.0000 | 1.0000 | 1.0000 | 0.0000 | 0.0000 |

Aggregate TP=22 FP=0 FN=0
Coverage complete: true
```

## Caveats

- Synthetic fixtures only; no live PII or credentials.
- Representative coverage, not exhaustive quality proof across all real-world shapes.
- Hardware and Go version affect timing; this report captures detection quality only.
