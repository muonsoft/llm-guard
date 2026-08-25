# D1 — Bilingual public README

Post-MVP documentation shipping boundary after archived PEB. OpenSpec change:
`bilingual-readme`.

## Outcome

GitHub visitors receive a concise English `README.md` and an equivalent Russian
`README.ru.md`. Both entry points explain the product boundary, quick start,
supported detector families, secure defaults, limitations, quality evidence,
and the route to deeper documentation without exposing milestone internals as
the primary navigation.

## Dependencies

PEB archived; public API, release-readiness evidence, security defaults, and
known limitations are already documented in repository sources.

## Out of scope

- Runtime, detector, policy, observer, or public Go API changes
- New performance or detection-quality claims
- Publishing the `v0.1.0` tag or GitHub release
- Translating the full `docs/` tree

## Verification

```bash
go test ./...
go vet ./...
go test -race ./...
./scripts/release-check.sh consumer
```

Review additionally checks language switching, link targets, command accuracy,
equivalent security/limitation statements, and that README examples follow the
tested public API in `example_test.go`.
