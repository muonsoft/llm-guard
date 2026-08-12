# Natasha reference harness

Development-only Python harness for reproducing pinned Natasha 0.8.0 extractor
output against synthetic PERSON and ADDRESS fixtures. Production Go builds and
tests never import this tooling.

## Pinned reference revisions

| Component | Revision |
|---|---|
| Python | `3.6.15@sha256:f8652afaf88c25f0d22354d547d892591067aa4026a7fa9a6819df9f300af6fc` |
| Natasha | `0.8.0`, commit `b603af32a598105d8d42c209a422436949a18e48` |
| Yargy | `0.9.0`, commit `c67041510d88981e19548715cd6fe6744d3e41e4` |
| Pymorphy2 | `0.8` |
| Pymorphy2 dictionaries | `2.4.393442.3710985` |

Exact hash-pinned transitive dependencies live in `requirements.lock`.

## CLI

```bash
python3 tools/natasha-reference/reference.py generate \
  --cases testdata/natasha/cases.jsonl \
  --output testdata/natasha/expected-python.jsonl

python3 tools/natasha-reference/reference.py verify \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl

python3 tools/natasha-reference/reference.py verify --offline \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl
```

- `generate` and live `verify` import the pinned Natasha extractors.
- `--offline` validates committed cases and expected output with stdlib only.
- Missing live dependencies produce an actionable install error; there is no
  silent offline fallback.

## Live Docker environment

Build the digest-pinned image and run live verification against PyPI:

```bash
docker build -f tools/natasha-reference/Dockerfile -t llm-guard-natasha-reference .

docker run --rm \
  -v "$PWD:/work" -w /work \
  llm-guard-natasha-reference verify \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl
```

Offline wheelhouse mirror for environments without PyPI access:

```bash
docker run --rm \
  -v "$PWD:/work" -v /tmp/r0-wheelhouse:/wheelhouse:ro -w /work \
  python:3.6.15@sha256:f8652afaf88c25f0d22354d547d892591067aa4026a7fa9a6819df9f300af6fc bash -lc \
  'python -m pip install --disable-pip-version-check --no-index --find-links=/wheelhouse --require-hashes -r tools/natasha-reference/requirements.lock && python tools/natasha-reference/reference.py verify --cases testdata/natasha/cases.jsonl --expected testdata/natasha/expected-python.jsonl'
```

Focused stdlib tests:

```bash
python3 -m unittest discover -s tools/natasha-reference -p 'test_*.py'
```

## JSONL schema version 1

### `testdata/natasha/cases.jsonl`

Each case record contains:

- `schema_version`: always `1`
- `id`: stable case identifier
- `entity`: `PERSON` or `ADDRESS`
- `corpus_class`: `positive`, `negative`, or `ambiguous`
- `input`: synthetic source text
- `product_expectation`: `match` or `no_match`
- `intentional_difference_class`: nullable `regression`,
  `intentional_difference`, or `unsupported_out_of_scope`
- `intentional_difference_reason`: nullable explanation when the class is set

### `testdata/natasha/expected-python.jsonl`

Each expected record binds the same `schema_version`, `id`, `entity`, and
`input` as its case and contains ordered `matches` in case order. Every match
includes:

- `entity`: must equal the case/expected entity
- `span`: Python codepoint half-open interval `{start, end}`
- `span_bytes`: converted UTF-8 byte half-open interval `{start, end}`
- `matched_text`
- `normalized`: Natasha fact fields reduced to JSON-safe values

Unknown fields, duplicate ids, invalid enums, invalid spans, entity drift,
record-order drift, and case/expected input drift fail validation. Offline
verify also requires canonical UTF-8 JSONL bytes for both cases and expected
files, including record order and final newline.

Output is canonical deterministic UTF-8 JSONL: one record per case, stable case
order, sorted object keys, final newline. Live generation run twice is
byte-identical.

## Codepoint to byte span conversion

Natasha/Yargy spans are Python codepoint offsets. Product Go spans use UTF-8 byte
offsets from `go-razdel`.

Conversion rule:

```python
byte_start = len(input[:codepoint_start].encode("utf-8"))
byte_end = len(input[:codepoint_end].encode("utf-8"))
```

Validation requires both:

- `input[codepoint_start:codepoint_end] == matched_text`
- `input.encode("utf-8")[byte_start:byte_end].decode("utf-8") == matched_text`

## Synthetic data policy

- Fixtures contain only synthetic Russian examples.
- No real contact data is committed.
- Cases document product expectations for M4/M5; they do not imply product
  detectors exist in R0.

## Precision-oriented metrics and difference classes

Metrics are reported independently per entity using the committed corpus
classes:

- exact-span true positives
- false positives and false negatives
- precision, recall, and exact-span rate when denominators are non-zero

Each Natasha/product divergence is classified as one of:

- `regression`: product behavior regressed relative to an accepted baseline
- `intentional_difference`: documented product safety boundary, for example M4
  rejecting isolated names that Natasha accepts
- `unsupported_out_of_scope`: reference behavior outside the bounded MVP matcher
  scope

Natasha output is diagnostic evidence, not product acceptance.
