## Why

После detection-only baseline библиотеке нужен первый полностью обратимый vertical slice, чтобы проверить публичный embedded workflow на реальном PII без зависимости от внешних сервисов. M1 переносит и уточняет EMAIL, resolver и masking/restore contracts из `docs/light_llm_guard_go_mvp_plan.md`, сохраняя pure Go и caller-owned state.

## What Changes

- Добавляется консервативный built-in EMAIL detector с UTF-8 byte spans и явными boundary rules.
- Добавляется общий детерминированный resolver для валидации, дедупликации и разрешения пересечений по внутреннему priority model.
- Добавляются `Mask`, `MaskResult`, opaque caller-owned `TokenSet` и collision-safe placeholders со случайным namespace.
- Добавляется точный `Restore`, раскрывающий только неизменённые placeholders из переданного `TokenSet`.
- Добавляются corpus, overlap, collision, repetition, Unicode, fuzz/property и embedded example проверки.

## Capabilities

### New Capabilities

- `structured-pii`: built-in EMAIL detection, boundary semantics и совместимость с общим detection pipeline.
- `finding-resolution`: детерминированная валидация, дедупликация, priority и non-overlap resolution findings.
- `reversible-masking`: caller-owned opaque tokens, collision-safe masking и exact restore semantics.

### Modified Capabilities

- Нет.

## Impact

- Расширяется публичный API корневого Go package: EMAIL detector registration, resolution, masking, restore и безопасные token types.
- Появляется только стандартная cryptographic randomness dependency; обязательных внешних сервисов и LLM/provider adapters нет.
- Existing detection-only API и канонический `core-detection` contract остаются совместимыми и concurrent-safe.
