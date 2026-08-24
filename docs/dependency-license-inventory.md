# Dependency and data license inventory

Complete supply-chain inventory for `github.com/muonsoft/llm-guard` release readiness.
Mechanical consistency is enforced by `./scripts/release-check.sh license` against
[dependency-license-modules.txt](dependency-license-modules.txt) and this file.

**Blocking rule:** any source without confirmed provenance and license evidence
blocks release until removed or documented here and, for shipped production Go
dependencies, in [THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES).

This inventory does **not** automatically adjudicate license compatibility; it
records evidence for maintainer review.

## Distribution topology

Three classes must not be conflated:

| Class | What it is | In llm-guard source module zip? | Runtime-linked in library? |
| --- | --- | --- | --- |
| **(a) Separate Go modules** | Dependencies resolved via `go get` / module proxy as their own modules | no (fetched separately) | production: `errors`, `go-razdel`; test-only: testify graph |
| **(b) Project-authored repository files** | Tables, corpora, harness source committed in this repo | yes (unless excluded by zip rules) | only compiled NLP tables/rules; testdata/tools are not linked at runtime |
| **(c) External reference-only tooling** | Python packages/wheels/dictionaries used offline by maintainers | no (not embedded) | no |

Direct Go dependency modules are **not** copied into this repository; they
download as separate modules. The llm-guard source zip **does** include
project-authored paths such as `testdata/` and `tools/natasha-reference/` even
when they are not runtime-linked.

## Go modules (`go list -m all`)

Evidence: committed `go.mod` / `go.sum` and module `LICENSE` files in the Go
module cache at audit time (2026-08-24). Production notices with full MIT text:
[THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES).

| Module | Version | License (evidence) | Role | Class | Source zip / runtime |
| --- | --- | --- | --- | --- | --- |
| `github.com/muonsoft/llm-guard` | (workspace) | MIT (`LICENSE`) | main library | (b) | yes / n/a |
| `github.com/muonsoft/errors` | v0.5.0 | MIT; Copyright (c) 2022 MuonSoft | direct runtime dependency | (a) | separate module / yes |
| `github.com/muonsoft/go-razdel` | v0.1.0 | MIT; Copyright (c) 2026 MuonSoft; ports Razdel `668dbe191a5cfd94bebf9155e2ffa5f94ff3fe33` | direct runtime tokenization | (a) | separate module / yes |
| `github.com/stretchr/testify` | v1.11.1 | MIT (module `LICENSE`) | test-only assertions | (a) | separate module / no |
| `github.com/stretchr/objx` | v0.5.2 | MIT (module `LICENSE`) | testify dependency | (a) | separate module / no |
| `github.com/davecgh/go-spew` | v1.1.1 | ISC (module `LICENSE`) | testify dependency | (a) | separate module / no |
| `github.com/pmezard/go-difflib` | v1.0.0 | BSD-3-Clause (module `LICENSE`) | testify dependency | (a) | separate module / no |
| `gopkg.in/yaml.v3` | v3.0.1 | Apache-2.0 / MIT (module `LICENSE`, `NOTICE`) | testify dependency | (a) | separate module / no |
| `gopkg.in/check.v1` | v0.0.0-20161208181325-20d25e280405 | BSD-style (module `LICENSE`) | indirect test helper | (a) | separate module / no |

Runtime direct dependency count remains **two** (`errors`, `go-razdel`).

### Upstream Razdel port notice

Pinned upstream revision `668dbe191a5cfd94bebf9155e2ffa5f94ff3fe33` (MIT;
Copyright (c) 2017 in pinned license text, no author name). Full notice in
[THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES).

## Project-authored tables, corpora, harness source (class b)

Present in the llm-guard **source module zip**; only compiled NLP tables/rules
are runtime-linked.

| Artifact | Path(s) | Provenance | License | Runtime-linked? | In source zip? |
| --- | --- | --- | --- | --- | --- |
| PERSON name/role tables | `internal/nlp/names.go`, related NLP tables | project-authored synthetic bounded lists | MIT (repository) | yes (compiled) | yes |
| ADDRESS labels/rules | `internal/nlp/address.go`, context helpers | project-authored rules | MIT (repository) | yes (compiled) | yes |
| Evaluation corpus | `testdata/evaluation/cases.jsonl` | synthetic fixtures | MIT (repository) | no | yes |
| Family corpora | `testdata/person/`, `testdata/address/`, `testdata/secrets/`, … | synthetic fixtures | MIT (repository) | no | yes |
| Natasha differential cases | `testdata/natasha/cases.jsonl` | synthetic PERSON/ADDRESS fixtures for reference comparison | MIT (repository) | no | yes |
| Natasha expected Python output | `testdata/natasha/expected-python.jsonl` | generated/verified reference spans (synthetic inputs) | MIT (repository) | no | yes |
| Natasha reference harness | `tools/natasha-reference/` | development-only Python harness; pinned revisions in README | MIT harness code; external Python deps per R0 | no | yes |
| Fuzz seeds | `fuzz_test.go` seed corpora | synthetic strings in tests | MIT (repository) | no | yes (via sources) |
| Secret pattern snapshot | [docs/secret-patterns.md](secret-patterns.md) | project-authored documentation | MIT (repository) | no | yes |

Tables are **not** populated from Natasha/OpenCorpora dictionary exports.

## External reference-only NLP tooling (class c)

Normative R0 decisions: [natasha-license-inventory.md](natasha-license-inventory.md).
These packages/wheels/dictionaries are **not** embedded in the Go module and are
not present in the llm-guard source zip as third-party artifacts.

| Source | License / provenance | Footprint | Decision |
| --- | --- | --- | --- |
| Natasha 0.8.0 | MIT code; bundled name lists lack standalone provenance | ~3.9 MB unpacked Python data | installed separately for harness; dictionaries **do not ship** |
| Yargy 0.9.0 | MIT | parser reference | development reference only |
| Pymorphy2 / OpenCorpora dicts | MIT code; CC BY-SA dictionary data | large external dicts | development reference only; **not embedded or redistributed** |
| Rejected Go morphology ports | various (see R0) | large embedded dicts | rejected for MVP |

The only approved third-party production NLP **Go module** is `go-razdel`.

## Maintainer update procedure

When `go.mod`, corpora, tables, or tooling changes:

1. Refresh `docs/dependency-license-modules.txt` from `go list -m all`.
2. Add a row here for every non-main module path and any new class-(b) artifact.
3. Update [THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES) for new **shipped**
   production Go dependencies (full MIT text when applicable).
4. Run `./scripts/release-check.sh license`.
5. Record evidence in the release checklist.
