## Why

Текущий корпус из 34 synthetic cases подтверждает только известный MVP-контракт и не показывает, сколько реальной или шумной PII остаётся вне него. Перед первым релизом нужно честно зафиксировать llm-guard как precision-oriented prefilter и получить воспроизводимые evidence его контрактной точности, границ покрытия и полного `Detect → Resolve → Mask → Restore` lifecycle без заявления high-recall/DLP-гарантий.

## What Changes

- Вводится модель оценки prefilter с раздельными профилями: обязательный offline contract/conformance gate и диагностический exposure benchmark на более широкой PII-разметке.
- Добавляется нормализованный versioned формат external/generated suites с provenance, immutable revision, SHA-256, source label, точными UTF-8 byte spans и явной классификацией `supported`, `unsupported` или `ignored` с причиной.
- Определяются воспроизводимые источники: pinned RedMadRobot RU PII holdout, FactRuEval для PERSON domain shift, ФИАС/ГАР как источник generated ADDRESS cases, checksum-aware generators для structured PII и synthetic/MIT fixtures для secrets.
- Расширяются отчёты: exact contract TP/FP/FN и rates остаются release gate; отдельно считаются sensitive-span coverage, leaked sensitive bytes, overmask и intentional/unsupported differences.
- Добавляется lifecycle-профиль для policy action, byte-for-byte `Mask → Restore`, placeholder integrity и безопасных диагностик, а также bounded performance evidence.
- Разделяются execution lanes: PR/release smoke остаётся offline и network-free; pinned external suite запускается явно либо по расписанию и привязывает отчёт к source manifest и commit.
- Публичная документация и release evidence явно позиционируют библиотеку как локальный precision-oriented prefilter, который снижает риск, но не заменяет high-recall DLP/NER и не гарантирует обнаружение любой PII.
- Изменение уточняет исходный MVP-план `docs/light_llm_guard_go_mvp_plan.md`: существующая консервативная PERSON/ADDRESS policy и запрет добиваться recall любой ценой становятся общей границей продукта и проверяемой evaluation policy.

## Capabilities

### New Capabilities

- `prefilter-evaluation`: versioned suites, source adapters/generators, contract/exposure/lifecycle metrics и правила воспроизводимого выполнения benchmark.

### Modified Capabilities

- `oss-distribution`: публичное позиционирование prefilter, известные ограничения и release evidence для offline и external evaluation lanes.

## Impact

- Затронуты `internal/evaluation`, `cmd/llmguard-eval`, evaluation fixtures/generators, documentation и CI/release scripts.
- Существующая schema v1 и строгий `-fail-on-regression` сохраняются как совместимый обязательный smoke gate; runtime detector behavior и public Go API не меняются.
- External datasets не становятся production dependency и не требуются обычному consumer или стандартному offline test run.
- Любые скачиваемые либо распространяемые data artifacts требуют pinned provenance/license manifest; реальные PII и действующие credentials не добавляются в repository, отчёты или logs.
