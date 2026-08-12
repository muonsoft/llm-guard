# PERSON quality report

Versioned black-box evaluation for the conservative RU `NewPersonDetector`.

## Corpus

| Field | Value |
|---|---|
| Corpus file | `testdata/person/cases.jsonl` |
| Schema version | `1` |
| Case count | 25 |
| Mandatory R0 positives | 8 |
| Mandatory R0 negatives | 4 |
| Extended boundary cases | 13 |

## Pinned production dependency

| Module | Pseudo-version | Commit |
|---|---|---|
| `github.com/muonsoft/go-razdel` | `v0.0.0-20260425122647-5cd53c7a1d02` | `5cd53c7a1d02780285406c6c9f1635a89953c27a` |

## Pinned reference baseline

| Component | Revision |
|---|---|
| Natasha | `0.8.0` / `b603af32a598105d8d42c209a422436949a18e48` |
| Yargy | `0.9.0` / `c67041510d88981e19548715cd6fe6744d3e41e4` |
| Reference fixtures | `testdata/natasha/cases.jsonl`, `testdata/natasha/expected-python.jsonl` |

## Metrics (mandatory corpus)

Evaluated offline by `TestPersonCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors` using span-set intersection (not per-case totals):

| Metric | Value |
|---|---|
| True positives | 13 |
| False positives | 0 |
| False negatives | 0 |
| Precision | 1.000 |
| Recall | 1.000 |
| Exact-span rate | 1.000 |

Counts reflect exact UTF-8 byte spans across all 25 corpus cases (13 expected PERSON spans, 0 erroneous matches).

## Intentional Natasha differences (R0 only)

| R0 case ID | Product policy | Classification |
|---|---|---|
| `person-negative-isolated-first-009` | Reject isolated first names | `intentional_difference` |
| `person-negative-isolated-surname-010` | Reject isolated surnames | `intentional_difference` |
| `person-negative-role-context-011` | Reject single-name role context | `intentional_difference` |
| `person-negative-street-like-012` | Reject street-like surname context | `intentional_difference` |

Product-only street regressions (`person-street-prospect-019`, `person-street-abbrev-020`) are not Natasha R0 differentials.

## Restore limitation

`Guard.Restore` returns the original masked substring literally. If an LLM moves a PERSON placeholder into a context that requires a different Russian case, the library does not inflect or re-agree the restored name.

## Verification commands

```bash
go test ./... -run 'Test(Person|Name|Morph|Token)'
go test -race ./... -run 'TestPerson'
go test ./... -run 'TestPersonCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors' -v
python3 tools/natasha-reference/reference.py verify --offline \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl
```
