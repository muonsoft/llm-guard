## Why

Текущий workflow `Release check (dry-run)` только проверяет tag и не создаёт
GitHub Release. Перед первым публичным выпуском нужно перенести проверенную модель
из соседнего `comm-relay`: явный maintainer dispatch после зелёного CI, повторная
валидация release candidate, финализация changelog, tag и публикация.

## What Changes

- Добавляется GitHub `Release` workflow с ручным SemVer input и поддержкой уже
  подготовленного SemVer tag.
- Публикация получает `contents: write` только в отдельной job после успешных
  offline и vulnerability gates; обычный CI остаётся read-only.
- Добавляется идемпотентная подготовка changelog: planned `v0.1.0` или будущий
  `[Unreleased]` превращается в датированную release section до создания tag.
- Release documentation переводится с полностью ручного tag/release процесса на
  явный maintainer-dispatched workflow.
- Из `docs/light_llm_guard_go_mvp_plan.md` уточняется финальная OSS shipping
  boundary: проверки остаются автоматическими, а решение начать публикацию —
  явным действием maintainer.

## Capabilities

### New Capabilities

Нет.

### Modified Capabilities

- `oss-distribution`: release-readiness расширяется проверяемой GitHub publication
  после зелёных gates, с точным SemVer, release commit и ограниченными permissions.

## Impact

Затрагиваются `.github/workflows/`, release scripts, `CHANGELOG.md`, README и
release/compatibility/readiness documentation. Public Go API, runtime module
graph и detector behavior не меняются. Сам change не выполняет push, tag или
publication.
