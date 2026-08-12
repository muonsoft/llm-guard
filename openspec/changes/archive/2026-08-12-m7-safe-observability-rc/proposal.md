## Why

Функциональный MVP уже обнаруживает, разрешает, маскирует и восстанавливает PII и secrets, но у caller нет безопасного framework-neutral способа наблюдать lifecycle операций и измерять качество полного detector pack. Этот change переносит и уточняет production concerns из `docs/light_llm_guard_go_mvp_plan.md`, чтобы библиотека достигла состояния release candidate без неявной утечки исходного текста, findings values или TokenSet mappings.

## What Changes

- Добавляется safe-by-default observer lifecycle для detect, mask и restore с Noop behavior без обязательного logging/metrics framework.
- Фиксируются stable production event schemas и outcomes для success, detector error, policy block и restore miss; разрешены только bounded lengths, durations и low-cardinality counts по entity/action.
- Добавляется отдельный conspicuous unsafe development profile, который невозможно включить неявно и который явно предупреждает о раскрытии original/masked text и raw diagnostic data.
- Добавляется pure-Go evaluation runner для общего annotated corpus с per-entity TP/FP/FN/TN, precision, recall, F1, FPR и FNR по всем MVP entities.
- Добавляются representative RU/mixed benchmarks, воспроизводимый baseline report и компилируемый embedded Detect → Mask → Restore example.
- README получает security considerations, observability contract, quality/benchmark commands и известные ограничения; проводится регрессионный audit всех logging, formatting и error surfaces.

## Capabilities

### New Capabilities

- `safe-observability`: Safe/unsafe observer profiles, stable lifecycle/metrics semantics, per-entity evaluation runner, benchmark baseline и release-candidate documentation boundary.

### Modified Capabilities

Нет: существующие detection, resolution, masking и policy semantics не меняются; observer только сообщает их результат через новый additive contract.

## Impact

- Публичный Go API расширяется immutable observer option, safe event/value types и явно опасным development-only diagnostics opt-in.
- Guard operations получают bounded duration/count instrumentation без logging dependency и без mutable global state.
- Добавляются reusable evaluation package/fixtures, benchmarks, quality baseline и README/package examples; generated ad-hoc reports остаются вне Git.
- Новые runtime dependencies, exporter server, Prometheus client, persistent audit store, OpenAI adapters и network calls не добавляются.
