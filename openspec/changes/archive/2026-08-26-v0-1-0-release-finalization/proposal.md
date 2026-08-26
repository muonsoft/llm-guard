## Why

Аудит перед первым релизом показал, что все существующие release gates проходят
на заявленном минимуме Go 1.26.2, хотя `govulncheck` на этой toolchain находит
достижимые уязвимости standard library. Перед `v0.1.0` нужно поднять minimum до
исправленного Go 1.26.6 и сделать security advisory check отдельным обязательным
release evidence, не разрушая offline dry-run и не публикуя release автоматически.

Change уточняет OSS-readiness часть исходного
`docs/light_llm_guard_go_mvp_plan.md`: проверяемая toolchain и отсутствие известных
достижимых уязвимостей становятся явной release boundary.

## What Changes

- Минимальная поддерживаемая и canonical release toolchain повышается с Go 1.26.2
  до Go 1.26.6 во всех executable и public contracts.
- `scripts/release-check.sh` получает отдельный pinned `vuln` mode; основной full
  dry-run остаётся network-free, а vulnerability gate явно требует network/cache.
- CI получает обязательный vulnerability job на exact minimum toolchain.
- Release checklist и readiness matrix требуют одновременно зелёный offline
  dry-run и зелёный vulnerability scan на release commit.
- Changelog и публичные README финализируются для предрелизного `v0.1.0`; tag,
  push и GitHub Release остаются отдельными maintainer actions.
- **BREAKING** изменений public Go API или runtime defaults нет; повышение
  минимальной patch-version Go является намеренным compatibility boundary.

## Capabilities

### New Capabilities

Нет.

### Modified Capabilities

- `oss-distribution`: release readiness теперь требует исправленную минимальную
  Go toolchain и отдельный воспроизводимый vulnerability gate, не смешанный с
  network-free dry-run.

## Impact

Затрагиваются `go.mod`, package documentation, bilingual README, compatibility и
release documents, `scripts/release-check.sh`, GitHub Actions и changelog. Новые
runtime dependencies не добавляются; pinned `govulncheck` запускается только как
release/CI tooling. Detection, masking, restore, policy и observability semantics
не меняются.
