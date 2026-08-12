## Why

После обратимого EMAIL vertical slice библиотеке не хватает практически полезного покрытия основных детерминированных PII, поэтому caller всё ещё вынужден поставлять собственные PHONE, IP, URL и numeric identifier detectors. M2 переносит и уточняет structured detector pack из `docs/light_llm_guard_go_mvp_plan.md`, сохраняя conservative matching, pure Go и общий `Detect → Resolve → Mask → Restore` pipeline.

## What Changes

- Добавляются built-in detectors для RU PHONE с conservative international baseline, IPv4/IPv6, URL, INN, SNILS и BANK_CARD.
- Для IP, INN, SNILS и BANK_CARD candidate matching дополняется semantic/checksum validation; длинные цифровые строки без подтверждающего формата не классифицируются агрессивно.
- Уточняются outer boundaries и overlap semantics для URL/EMAIL и structured numeric entities.
- Расширяется internal resolver priority model для новых entity types без появления публичной policy-конфигурации.
- Добавляются positive, negative, malformed, Unicode, mixed overlap, end-to-end и per-entity evaluation tests.

## Capabilities

### New Capabilities

- Нет.

### Modified Capabilities

- `structured-pii`: расширение built-in coverage от EMAIL до основного PHONE/IP/URL/INN/SNILS/BANK_CARD pack с формализованными validators и boundaries.
- `finding-resolution`: deterministic priorities и overlap semantics для EMAIL, URL и structured numeric findings.

## Impact

- Расширяется additive public API корневого Go package новыми immutable detector constructors и entity constants.
- Existing `Guard`, `WithDetector`, `Resolve`, `Mask` и `Restore` остаются общим pipeline без entity-specific branching.
- Runtime остаётся pure Go и offline; для semantic parsing используются стандартная библиотека и локальные checksum algorithms без network calls или новых внешних зависимостей.
