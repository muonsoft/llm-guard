# ADDRESS quality report

Versioned black-box evaluation for the conservative RU `NewAddressDetector`.

## Corpus

| Field | Value |
|---|---|
| Corpus file | `testdata/address/cases.jsonl` |
| Schema version | `1` |
| Case count | 33 |
| Mandatory R0 positives | 5 |
| Mandatory R0 negatives | 3 |
| Mandatory R0 ambiguous | 1 |
| Extended boundary cases | 24 |

## Composition matrix

| Parts | Result |
|---|---|
| settlement / region / street / house alone | reject |
| settlement + street (no house) | reject |
| street + house | accept |
| settlement + street + house | accept maximal span |
| street + house + corpus / building / apartment | accept maximal span |

Postal index, geocoding, FIAS normalization, and generic location NER are out of scope for M5.

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

Evaluated offline by `TestAddressCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors` using span-set intersection:

| Metric | Value |
|---|---|
| True positives | 20 |
| False positives | 0 |
| False negatives | 0 |
| Precision | 1.000 |
| Recall | 1.000 |
| Exact-span rate | 1.000 |

Counts reflect exact UTF-8 byte spans across all 33 corpus cases (20 expected ADDRESS spans, 0 erroneous matches).

## Intentional Natasha differences (R0 only)

R0 ADDRESS fixtures align with the conservative product grammar on pinned Natasha output; no unexplained R0 differentials are expected for M5.

## Restore limitation

`Guard.Restore` returns the original masked substring literally. The library does not normalize, inflect, or geocode restored addresses.

## Verification commands

```bash
go test ./... -run 'Test(Address|Addr|Resolve.*Address|Mixed)' -count=1
go test -race ./... -run 'TestAddress' -count=1
go test ./... -run 'TestAddressCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors' -v -count=1
PYTHONDONTWRITEBYTECODE=1 python3 tools/natasha-reference/reference.py verify --offline \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl
```
