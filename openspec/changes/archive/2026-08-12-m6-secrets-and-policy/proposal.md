## Why

Текущий Guard защищает structured PII, но не распознаёт credentials и всегда трактует найденный span как подлежащий маскированию. Следующий MVP-срез переносит из `docs/light_llm_guard_go_mvp_plan.md` базовое обнаружение secrets и уточняет минимальную action model, чтобы вызывающий код мог безопасно маскировать либо блокировать запрос без внешнего policy engine.

## What Changes

- Добавляются conservative deterministic detectors для structurally valid JWT, PEM private-key blocks, version-pinned GitHub/GitLab/OpenAI/AWS-like tokens и credential-bearing DSN.
- Добавляются строковые actions `allow`, `mask`, `block` и immutable per-entity/per-family configuration с однозначными safe defaults: PII маскируется, secrets блокируются, если caller явно не выбрал mask или allow.
- `Guard.Mask` применяет action после общего detection/resolution: allow исключает finding из замены, mask создаёт reversible token, block возвращает отдельную проверяемую safe error и zero result.
- Resolver получает deterministic priority для secrets, включая вытеснение URL/EMAIL fragments внутри credential-bearing DSN.
- Добавляются versioned corpus, malformed/negative cases, overlap, safe-formatting, cancellation и concurrency tests; документируются источники pattern versions и ограничения.

## Capabilities

### New Capabilities

- `secret-detection`: Поддерживаемые secret families, structural validation, versioned patterns, безопасные metadata и corpus boundaries.
- `minimal-policy`: Action model, default и configured actions, block semantics и безопасное concurrent применение policy.

### Modified Capabilities

- `finding-resolution`: Secret entities получают deterministic overlap priorities относительно URL, EMAIL и generic findings.
- `reversible-masking`: Mask применяет allow/mask/block policy и не возвращает partial либо misleading masked result при block.

## Impact

- Публичный Go API расширяется новыми `EntityType`, constructors встроенных detectors, `Action`, options конфигурации и проверяемой block error.
- Меняются core Guard pipeline, resolver priority table и masking selection; существующий default PII masking и restore contract остаются совместимыми.
- Добавляются pure-Go regexp/parsing implementation, corpus fixtures и документация без network calls, provider SDK, persistence, logging либо policy-framework dependencies.
