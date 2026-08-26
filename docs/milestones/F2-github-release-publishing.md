# F2 — GitHub release publishing

## Shipping boundary

`llm-guard` получает maintainer-dispatched GitHub Release workflow, повторяющий
проверенную модель соседнего `comm-relay`: после зелёного CI maintainer задаёт
SemVer, workflow повторно валидирует release candidate, финализирует changelog на
`main`, создаёт tag на release-коммите и публикует source-only GitHub Release.

## Dependencies

- F1 archived and reviewed green.
- No active OpenSpec change at session start.
- Herdr variant C preflight passes.

## Scope

- Replace the dry-run-only tag workflow with an explicit release publication
  workflow for manual dispatch and already-prepared tag pushes.
- Add an idempotent changelog preparation/check script for planned first-release
  and future `[Unreleased]` sections.
- Keep normal CI read-only and require the release workflow to rerun the full
  offline and exact-minimum vulnerability gates before publication.
- Make public release documentation consistent with the automated workflow.

## Acceptance

1. Manual `Release` dispatch from current `main` with a valid SemVer finalizes
   release metadata, tags the resulting commit, and creates a GitHub Release only
   after validation succeeds.
2. A pushed SemVer tag can publish only when its changelog section is already
   finalized; invalid versions, stale/non-main dispatches, and planned changelog
   state fail before publication.
3. Workflow permissions are read-only except for the publication job, input is
   not interpolated as shell code, and concurrent releases are serialized.
4. Local tests prove changelog check/apply/idempotency and GitHub Actions syntax;
   broad release, vulnerability, Go, and OpenSpec gates stay green.

## Verification boundary

```bash
bash -n scripts/prepare-release.sh
bash scripts/prepare-release.sh 0.1.0 --check-only
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
openspec validate github-release-publishing --strict --no-interactive
openspec validate --specs --strict --no-interactive
```

## Non-goals

- No push, tag, workflow dispatch, or GitHub Release in this milestone.
- No binary archives: Go consumers install the module from its semantic tag.
- No change to detector behavior or public Go API.
