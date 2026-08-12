## Why

После основного structured pack библиотеке не хватает оставшихся PII-сущностей MVP и безопасного regexp-adapter для пользовательских строковых entity. M3 закрывает structured scope перед NLP-milestones и уточняет консервативные границы из `docs/light_llm_guard_go_mvp_plan.md`, чтобы обычные даты и реквизиты не маскировались как PII без достаточных оснований.

## What Changes

- Добавляются консервативные встроенные detectors для российского паспорта, банковского счёта и даты рождения с явным RU-контекстом.
- Определяются точно поддерживаемые формы, normalization/validation и negative corpus для неоднозначных цифровых строк, обычных дат и договорных реквизитов.
- Добавляется публичный immutable `CustomRegexpDetector` для string-based custom entity с безопасной проверкой имени, entity, pattern, confidence, zero-width matches и UTF-8 boundaries.
- Подтверждается совместимость regexp и custom Go detectors с общими `Detect`, `Resolve`, `Mask`, `Restore`, priorities и concurrent use.
- Проводится public API review M0–M3 и фиксируется стабильная boundary для последующих MVP milestones без запланированных breaking core changes.

## Capabilities

### New Capabilities

- `custom-detection`: публичный regexp-adapter и end-to-end contract для string entity и пользовательских Go detectors.

### Modified Capabilities

- `structured-pii`: добавляются PASSPORT, BANK_ACCOUNT и contextual DATE_OF_BIRTH, их точные conservative matching rules, corpus и общий reversible pipeline.

## Impact

Изменяются корневой публичный Go API, built-in structured detectors, resolver priorities, structured corpus/tests, examples и документация public API. Новые runtime dependencies и внешние сервисы не добавляются; реализация остаётся pure Go и embeddable.
