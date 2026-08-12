# R0 verification evidence

## Reviewed outcome

- Architecture: external `github.com/muonsoft/go-razdel`, internal bounded
  product matcher, project-owned token annotations/tables, no MVP `go-natasha`
  and no production morphology dictionary.
- Reference: Natasha 0.8.0 / Yargy 0.9.0 / Pymorphy2 0.8 under digest-pinned
  Python 3.6.15, with complete hash-pinned transitive lock.
- Corpus: 21 synthetic PERSON/ADDRESS cases and deterministic schema-v1 expected
  output with Python codepoint and UTF-8 byte spans.
- Production dependency baseline: `go.mod`, `go.sum`, `go list -deps ./...`, and
  `go mod graph` contain no reference Python/NLP dependency and are unchanged by
  R0 tooling.

## Composer and independent review

- Primary result: `.agent-orchestration/results/r0-primary.md`.
- Consolidated correction: `.agent-orchestration/results/r0-correction.md`.
- Verification-evidence correction: `.agent-orchestration/results/r0-correction2.md`.
- Orchestrator reviewed every changed harness/fixture file and reran focused,
  live-reference and broad checks. Review corrections added strict entity/order/
  canonical schema binding, digest pinning, schema-error containment and
  generator self-validation.

## Reference and focused checks

```bash
python3 -m unittest discover -s tools/natasha-reference -p 'test_*.py' -v
# PASS: 27 tests

python3 tools/natasha-reference/reference.py verify --offline \
  --cases testdata/natasha/cases.jsonl \
  --expected testdata/natasha/expected-python.jsonl
# PASS

docker run --rm \
  -v "$PWD:/work" -v /tmp/r0-wheelhouse:/wheelhouse:ro -w /work \
  python:3.6.15@sha256:f8652afaf88c25f0d22354d547d892591067aa4026a7fa9a6819df9f300af6fc \
  bash -lc 'python -m pip install --disable-pip-version-check --no-index --find-links=/wheelhouse --require-hashes -r tools/natasha-reference/requirements.lock -q && python tools/natasha-reference/reference.py generate --cases testdata/natasha/cases.jsonl --output testdata/natasha/expected-python.jsonl && python tools/natasha-reference/reference.py verify --cases testdata/natasha/cases.jsonl --expected testdata/natasha/expected-python.jsonl && python tools/natasha-reference/reference.py generate --cases testdata/natasha/cases.jsonl --output /tmp/gen-a.jsonl && python tools/natasha-reference/reference.py generate --cases testdata/natasha/cases.jsonl --output /tmp/gen-b.jsonl && cmp -s /tmp/gen-a.jsonl /tmp/gen-b.jsonl && cmp -s /tmp/gen-a.jsonl testdata/natasha/expected-python.jsonl'
# PASS: live verify and two-generation byte identity
```

The optional normal-PyPI Dockerfile build was canceled after exceeding one
minute in this environment's restricted Docker network. This is not a product or
reference blocker: the same digest-pinned interpreter and checked-in hash lock
passed the mandatory offline-wheelhouse live extraction above.

## Broad and OpenSpec checks

```bash
go list -deps ./...
# PASS; production packages unchanged

go mod graph
# PASS; production module graph unchanged

go test ./...
# PASS: ok github.com/muonsoft/llm-guard

go vet ./...
# PASS

go test -race ./...
# PASS: ok github.com/muonsoft/llm-guard

openspec validate r0-nlp-decision-gate --type change --strict --no-interactive
# PASS: Change 'r0-nlp-decision-gate' is valid

openspec validate --specs --strict --no-interactive
# PASS: 1 passed, 0 failed
```

## Blockers

None. M4/M5 may consume the approved boundary without reopening repository,
tokenizer, parser-family, or production morphology-data decisions.

## OpenSpec verification report

| Dimension | Status |
|---|---|
| Completeness | PASS — 29/29 tasks and 7/7 requirements covered |
| Correctness | PASS — all 13 scenarios mapped to reviewed evidence |
| Coherence | PASS — implementation follows the six design decisions |

Requirement mapping:

1. **Complete reproducible dependency audit:** exact revisions, source commands,
   transitive extractor graphs, primitives and grammemes are in
   `docs/natasha-port-scope.md`; unresolved Natasha list provenance is excluded,
   not assumed.
2. **NLP runtime boundary:** `docs/adr/0003-nlp-runtime-boundary.md` fixes external
   `go-razdel`, internal product runtime, no `go-natasha`, and a gated future
   extraction criterion.
3. **Bounded parser strategy:** the closed feature matrix proves every mandatory
   M4/M5 form is finite; upstream-only relations/interpretation/recursion remain
   out of scope. No blocking fixture was found.
4. **Licensed morphology/dictionaries:** `docs/natasha-license-inventory.md`
   records revision, license/provenance, footprint/runtime properties and
   distribution decision for every audited source; production uses authored
   bounded annotations with no imported data.
5. **Separated reference harness:** `tools/natasha-reference/` is isolated,
   digest/hash pinned and live reproducible; 27 tests cover schema/offset/error
   invariants; ordinary Go checks pass without Python or downloads.
6. **Intentional product differences:** the schema and 21 cases distinguish
   corpus/product expectation from Natasha output. Isolated PERSON cases retain
   reference matches but are marked intentional; standalone ADDRESS parts retain
   empty reference matches and product `no_match` policy.
7. **M4/M5-ready exit evidence:** ADR, audit, license inventory, harness README,
   committed fixtures and this command ledger close repository, tokenizer,
   parser-family and data-source decisions.

Issues: no CRITICAL, WARNING, or SUGGESTION findings. Ready for spec sync and
archive.
