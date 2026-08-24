# PEB — Prefilter evaluation benchmark

Post-MVP shipping boundary after archived M8. OpenSpec change:
`prefilter-evaluation-benchmark`.

## Outcome

llm-guard is evaluated as a precision-oriented prefilter: strict offline
conformance stays on schema v1, while versioned suite v2 measures contract,
exposure, and Mask/Restore lifecycle without claiming high-recall DLP.

## Dependencies

M8 archived (OSS `v0.1.0` repository readiness).

## Out of scope

- Detector syntax expansion, single-token PERSON, incomplete ADDRESS, ML/NER
- Public evaluation SDK
- Vendoring raw external corpora into the Go module
- Network access from `go test ./...` or the evaluator

## Verification

```bash
go test ./...
go vet ./...
go test -race ./...
go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format markdown -fail-on-regression
```

External fetch is explicit-only and is not part of the default PR gate.
