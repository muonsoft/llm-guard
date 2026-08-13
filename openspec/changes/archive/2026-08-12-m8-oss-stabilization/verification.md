# Verification Report: m8-oss-stabilization

## Summary

| Dimension | Status |
|---|---|
| Completeness | 36/36 tasks complete; 6/6 requirements implemented |
| Correctness | 14/14 scenarios covered by executable checks, CI/config review and durable evidence |
| Coherence | Design decisions followed; no unresolved contradiction or project-pattern deviation |

## Completeness

### Task coverage

- `openspec instructions apply --change m8-oss-stabilization --json` reports
  `36/36`, state `all_done`.
- Public module boundary, CI/release automation, dependency/data audit, OSS
  policies, readiness ledgers and every required verification command are present.
- `openspec validate m8-oss-stabilization --strict --no-interactive` and
  `git diff --check` pass after all review corrections.

### Requirement coverage

1. **Внешнее подключение и package documentation** — `doc.go` defines canonical
   import/toolchain, concurrency, caller-owned state and security boundaries;
   `testdata/external-consumer/main.go:13` is a standalone public-only fixture and
   explicitly executes Detect (`:26`), Mask (`:36`) and Restore (`:42`).
   `scripts/release-check.sh:52` creates and cleans an independent temporary module.
2. **Public API и compatibility policy** — `docs/compatibility-versioning.md`
   publishes the pre-1.0 SemVer/API/Go policy, while `README.md:5`, `README.md:219`
   and `docs/known-limitations.md` publish the pre-release boundary and complete
   MVP limitations. M8 changes GoDoc only; exported runtime symbols and
   `go.mod`/`go.sum` remain unchanged.
3. **Воспроизводимые CI quality gates** — `.github/workflows/ci.yml:14` runs
   test/vet on Go `1.26.2` and `stable`, with separate race, consumer,
   evaluation, exact fuzz and license jobs. `scripts/release-check.sh:159` runs
   all three required bounded fuzz targets; `:185` composes the full dry-run.
4. **Лицензии, notices и provenance** — `THIRD_PARTY_NOTICES` contains complete
   MIT notices for `muonsoft/errors`, `go-razdel` and pinned upstream Razdel;
   `docs/dependency-license-inventory.md` separates Go modules, source-module
   contents and non-embedded reference tooling. `scripts/release-check.sh:85`
   compares `go list -m all`, the exact path/version inventory rows and notice
   markers. Negative tests prove missing/stale evidence fails.
5. **Публичные security и contribution policies** — `SECURITY.md` gives private
   reporting and synthetic-only disclosure rules; `CONTRIBUTING.md` requires
   OpenSpec, targeted tests and provenance/license evidence for detector/data or
   dependency changes.
6. **Release dry-run и readiness evidence** — `.github/workflows/release-check.yml`
   has `contents: read` only and runs the full dry-run manually/on version-like
   tags without publishing. `docs/release-checklist.md` keeps tag/push/release as
   separate manual actions; `docs/mvp-readiness-matrix.md` maps every MVP DoD
   item; `docs/m8-quality-benchmark-comparison.md` compares evaluation and stable
   benchmark names to M7 with hardware/no-SLO caveats.

## Correctness

All 14 scenarios have direct evidence:

- an independent temporary module imports only the canonical path and runs
  explicit Detect → Mask → Restore on synthetic data;
- compatibility policy explains pre-1.0 change handling, minimum Go and all MVP
  boundaries without claiming that `v0.1.0` is already published;
- CI covers minimum/stable test+vet and independent race/consumer/evaluation/fuzz/
  license gates without an external service;
- exact-name bounded fuzz and the evaluator's non-zero regression behavior are
  invoked by the release script;
- every Go module path/version and source/data class has provenance/license and
  distribution-role evidence, with full required production notices and a
  blocking consistency gate;
- vulnerability reports use private/synthetic guidance and contributions that
  change data/dependencies require provenance review;
- the full dry-run performs no tag, push, upload, publication or release mutation,
  and the checklist reserves those operations for explicit maintainer action;
- quality remains TP=22, FP=0, FN=0 with complete coverage for all 16 MVP
  entities, and all 11 stable benchmark names execute successfully.

Independent verification commands:

- `./scripts/release-check.sh` — PASS: gofmt/diff hygiene, `go test ./...`,
  `go vet ./...`, `go test -race ./...`, external consumer, evaluation, exact
  inventory/notices, 3×2s fuzz and benchmark smoke.
- Evaluator inside the dry-run — PASS: 34 cases, TP=22, FP=0, FN=0, all 16
  entities have positive and negative coverage.
- Benchmarks inside the dry-run — PASS: 11/11 stable Detect/Mask/Restore/Observer
  names.
- Negative inventory evidence — PASS: a missing module path and a stale
  `testify` version both cause non-zero license-gate results.
- `openspec validate m8-oss-stabilization --strict --no-interactive` — PASS.
- `git diff --check` — PASS.
- Post-fix `/tmp/llm-guard-fuzz.*` check — no orphaned cache directories.

## Coherence

- M8 adds distribution/config/documentation surfaces only; runtime public API,
  dependency graph, safe defaults, byte spans, concurrency and TokenSet ownership
  are unchanged.
- One POSIX script is the shared local/CI command contract; expensive race/fuzz/
  benchmark gates are not duplicated across the moving Go matrix.
- The external consumer uses a transient `go.mod` plus local `replace`; no nested
  module or unpublished-version fetch is committed.
- Supply-chain documentation distinguishes source-module presence from runtime
  linkage and external reference packages, avoiding claims that Natasha or
  OpenCorpora dictionaries ship.
- Release readiness and release publication remain separate. Workflows have
  read-only repository permissions and contain no upload/publish step.
- The one-line orchestrator cleanup correction isolates fuzz `GOCACHE` in a
  subshell, so the following benchmark cannot recreate an orphaned temporary
  directory.

## Issues by priority

### CRITICAL

- None.

### WARNING

- None.

### SUGGESTION

- None.

## Orchestration evidence

- Variant C used one Composer 2.5 session for one primary job and two correction
  jobs in the same milestone/worktree.
- Whole-diff review corrected incomplete third-party notices, source-module
  distribution terminology, exact module/version inventory enforcement,
  external Detect coverage and pre-publication changelog links.
- Independent verification revealed and closed stale-version gate behavior and a
  temporary fuzz-cache cleanup issue before acceptance.
- Primary/correction result files carry all declared completion markers. Final
  bounded Herdr wait/get/read calls completed; transport classified `healthy`,
  with no orphaned pane and no control-call timeout after the initial approval
  review delay.

## Final assessment

All tasks, requirements, scenarios, supply-chain evidence, external consumer
boundary, OSS policies, CI/dry-run gates and release-readiness documentation are
green. No CRITICAL or WARNING remains. Ready for spec sync and archive.
