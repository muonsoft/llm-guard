## Why

Библиотеке нужен минимальный полезный detection-only core, прежде чем добавлять встроенные PII-детекторы и обратимую маскировку. Этот change переносит и уточняет Phase 0 из `docs/light_llm_guard_go_mvp_plan.md`, чтобы внешний Go-код уже мог безопасно подключать собственные детекторы.

## What Changes

- Добавляется компактный публичный API для сущностей, findings, detector-расширений и immutable `Guard`.
- Добавляется конкурентное выполнение пользовательских детекторов с детерминированной агрегацией результатов и явной обработкой отмены контекста и ошибок.
- Добавляется строгая проверка UTF-8 byte spans, confidence и безопасных detector metadata без включения исходного текста в ошибки.
- Добавляются тесты публичного API, Unicode spans, failure semantics и concurrent use, а также минимальная CI-проверка.
- Фиксируются архитектурные решения о byte offsets и stateless core с caller-owned future state.

## Capabilities

### New Capabilities

- `core-detection`: Публичный detection-only core, контракт custom detector, валидация findings, конкурентное выполнение и детерминированный результат.

### Modified Capabilities

Нет.

## Impact

- Новый корневой Go package модуля `github.com/muonsoft/llm-guard` и его публичный API.
- Новые package/example/race tests без provider, proxy, logging или persistence интеграций.
- Новые ADR в `docs/adr/` и базовый GitHub Actions workflow для `go test`, `go vet` и race-проверки.
- Следующие milestone смогут добавлять built-in detectors и masking поверх стабильного detection-контракта.
