# Verification — github-release-publishing

## Outcome

**PASS** — все 12 tasks завершены; CRITICAL и WARNING отсутствуют. Реализация
удовлетворяет delta requirement и design. Ни commit/push/tag, ни workflow dispatch,
ни GitHub Release в ходе verification не выполнялись.

## Requirement and scenario evidence

- **Dry-run безопасен по умолчанию:** `GOTOOLCHAIN=go1.26.6
  ./scripts/release-check.sh` PASS; script не меняет refs и не публикует artifacts.
- **Security gate отделён:** `GOTOOLCHAIN=go1.26.6
  ./scripts/release-check.sh vuln` PASS, `No vulnerabilities found`.
- **Release выполняется только отдельно:** `.github/workflows/release.yml` имеет
  manual `version` input, `validate → publish`, current-main guard, changelog-only
  release commit и точный `target_commitish`.
- **Validation блокирует публикацию:** strict SemVer, changelog readiness,
  full/vuln gates и `needs: validate` предшествуют write job.
- **Подготовленный tag не переписывает history:** tag-push path требует
  `--require-final`; workflow сравнивает peeled `^{commit}` для lightweight и
  annotated tags и не содержит tag move/delete.
- **Permissions ограничены:** top-level `contents: read`; только publish job имеет
  `contents: write`. Release input передаётся через environment и проходит strict
  SemVer до записи в outputs.
- **M7 baseline:** существующий quality/benchmark report и обязательные offline
  gates сохранены; external RedMadRobot evidence явно optional diagnostic и не
  приписывается changelog-only tag commit.

## Independent commands

| Check | Result |
| --- | --- |
| `bash -n scripts/prepare-release.sh scripts/prepare-release-test.sh` | PASS |
| `bash scripts/prepare-release-test.sh` | PASS, 16 passed / 0 failed |
| `bash scripts/prepare-release.sh 0.1.0 --check-only` | PASS |
| invalid SemVer and literal-header regression demonstrations | PASS (rejected) |
| temporary lightweight/annotated tag peeling via `^{commit}` | PASS |
| `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7` | PASS |
| `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh` | PASS |
| `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln` | PASS, no vulnerabilities |
| `git diff --check` | PASS |
| `openspec validate github-release-publishing --strict --no-interactive` | PASS |

Full dry-run evidence includes tests, vet, race, external consumer, schema-v1 and
generated contract/lifecycle evaluation, dependency/license inventory, three
bounded fuzz targets and benchmark smoke.

## External state checked read-only

- GitHub recognizes current llm-guard workflows as active before this push.
- Repository Actions are enabled; current workflow-token policy permits write.
- `main` currently has no branch protection, so the documented bot changelog push
  path is viable today. Future branch protection without bot bypass will fail
  safely before tag creation, as documented in design.
- A successful CommRelay `Release` workflow_dispatch run was inspected as the
  behavioral reference; llm-guard intentionally omits its desktop build artifacts.

## Remaining release-time checks

- Hosted CI and the new Release workflow can only be observed after these changes
  are pushed to `main`.
- Maintainer must wait for green CI, then explicitly dispatch `Release` with
  `v0.1.0`; this milestone does not authorize that external mutation.
- After publication, verify clean-module `go get ...@v0.1.0`.

## Suggestion (non-blocking)

Consider pinning third-party `softprops/action-gh-release` to a reviewed full
commit SHA in a future supply-chain hardening change; current `@v2` matches the
proven CommRelay workflow and repository policy does not require SHA pinning.
