# Contributing to llm-guard

Thank you for helping improve llm-guard. This project uses OpenSpec for
spec-driven changes and expects conservative, test-backed contributions.

## Before you start

1. Read [AGENTS.md](AGENTS.md) for repository conventions.
2. For non-trivial work, propose or extend an OpenSpec change under
   `openspec/changes/` before implementing behavior changes.
3. Check [docs/compatibility-versioning.md](docs/compatibility-versioning.md)
   for pre-1.0 API expectations.

## Development setup

```bash
go test ./...
go vet ./...
```

Broader release gates live in `./scripts/release-check.sh` and
[docs/release-checklist.md](docs/release-checklist.md).

## Pull request expectations

- Keep changes focused; avoid unrelated refactors.
- Add or update tests for behavior you change.
- Run targeted tests locally before opening a PR.
- Ensure `go test ./...` and `go vet ./...` pass.
- For detector, corpus, dictionary, or dependency changes, update license/provenance
  evidence (see below).

## Tests and fixtures

- Prefer synthetic fixtures; never commit real PII, credentials, or private contact
  data.
- Family corpora live under `testdata/<entity>/cases.jsonl`; follow existing JSONL
  shape and UTF-8 byte-span conventions.
- Fuzz targets must remain bounded in CI smoke (`./scripts/release-check.sh fuzz`).
- Evaluation changes should keep `go run ./cmd/llmguard-eval ... -fail-on-regression`
  green unless an intentional baseline update is documented.

## Detector and data provenance

When changing detectors, tables, corpora, or dependencies:

1. Record source, license, and distribution role in
   [docs/dependency-license-inventory.md](docs/dependency-license-inventory.md).
2. Update [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES) when shipped production
   dependencies change.
3. Refresh [docs/dependency-license-modules.txt](docs/dependency-license-modules.txt)
   if `go.mod` / `go.sum` changes (`go list -m all`).
4. Do **not** copy Natasha, OpenCorpora, or other reference dictionaries into
   production tables. See [docs/natasha-license-inventory.md](docs/natasha-license-inventory.md).

Unknown sources or licenses block release readiness until resolved.

## OpenSpec workflow

Planning artifacts:

1. `proposal.md`
2. `specs/<capability>/spec.md`
3. `design.md`
4. `tasks.md`

Use the OpenSpec CLI (`openspec` 1.8+) or Cursor `/opsx-*` commands documented in
[AGENTS.md](AGENTS.md).

## Code style

- Pure Go, library-first design; no mandatory external services in core.
- Match existing naming, error handling (`github.com/muonsoft/errors`), and test
  patterns in neighboring files.
- Run `gofmt` on changed Go files.

## Security

Read [SECURITY.md](SECURITY.md). Do not disclose sensitive values in public
issues or PR descriptions.

## License

By contributing, you agree that your contributions are licensed under the
project MIT License.
