## Why

Функциональный MVP уже достиг release-candidate boundary в M7, но ещё не имеет
полного проверяемого OSS-контракта: внешнего consumer compile gate, законченного
license/provenance пакета, воспроизводимых CI/release checks и публичных policy и
release документов. M8 переносит Phase 6 и финальный Definition of Done из
`docs/light_llm_guard_go_mvp_plan.md` в каноническую готовность `v0.1.0`, не
публикуя сам release.

## What Changes

- Ввести capability `oss-distribution` с требованиями к поддерживаемой Go
  toolchain, внешнему подключению module, CI, bounded fuzz smoke и безопасному
  release dry-run.
- Завершить package/public API documentation review, executable examples и
  compile check из внешнего Go module.
- Зафиксировать compatibility/versioning policy, changelog и отдельный release
  checklist без автоматического tag, push или publish.
- Провести inventory лицензий production/test/tooling dependencies и provenance
  словарей/data; добавить требуемые third-party notices.
- Опубликовать `SECURITY.md`, `CONTRIBUTING.md`, финальную DoD/limitations matrix
  и regression comparison к M7 quality/benchmark baselines.
- Уточнить GitHub Actions так, чтобы обычный CI и release dry-run воспроизводили
  заявленные quality gates детерминированно и без publication side effects.

## Capabilities

### New Capabilities

- `oss-distribution`: проверяемый контракт подключения, совместимости,
  supply-chain disclosure, CI и безопасной подготовки OSS release.

### Modified Capabilities

- Нет. M8 стабилизирует distribution/release boundary и не меняет требования к
  detection, masking, policy или observability.

## Impact

Изменения затронут package documentation и examples, `.github/workflows/`,
release/security/contribution документы, dependency notices и финальные quality
reports. Public Go API и runtime dependency graph должны остаться совместимыми;
provider adapters, proxy, persistence, ML/NER и фактическая публикация release
остаются вне scope.
