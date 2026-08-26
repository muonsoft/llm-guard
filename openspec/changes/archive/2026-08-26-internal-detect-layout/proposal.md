## Why

Корневой пакет `llmguard` смешивает компактный публичный API с телами built-in detectors и пакетными хелперами. После MVP это мешает читать контракт библиотеки и скрывает, что большая часть кода не предназначена внешним consumer. Нужно разнести приватную логику в `internal/` без смены поведения.

Уточнение относительно `docs/light_llm_guard_go_mvp_plan.md` §19: исходный набросок предлагал публичные пакеты `detector/`, `resolver/`, `masking/`. Этот change сохраняет один корневой import path и прячет реализацию под `internal/`, как уже сделано для NLP.

## What Changes

- Тела structured PII и secret detectors переносятся в `internal/detect` и возвращают byte spans, а не публичные `Finding`.
- Общие хелперы границ, цифр и bounded RU-context переезжают в тот же пакет.
- Корневой пакет оставляет constructors (`NewEmailDetector` и остальные), `Detector`, `Guard` и тонкие адаптеры `Span → Finding`.
- PERSON/ADDRESS остаются вызовами `internal/nlp`, но получают тот же адаптерный слой, что и structured detectors.
- Публичные тесты, examples, fuzz targets и `internal/evaluation` не меняют контракт.

Это pure refactor: runtime поведение, exported API и default security policy не меняются. **BREAKING** изменений нет.

## Capabilities

### New Capabilities

- Нет. Поведение библиотеки не меняется; layout не является отдельной product capability.

### Modified Capabilities

- Нет. Требования `core-detection`, `structured-pii`, `secret-detection`, `russian-person`, `russian-address` и `oss-distribution` остаются в силе. Внешний consumer по-прежнему импортирует только `github.com/muonsoft/llm-guard`.

Change объявляет `skip_specs: true`.

## Impact

Затрагиваются корневые файлы detectors/helpers, новый пакет `internal/detect`, тонкий `builtin.go`, whitebox fuzz helpers и ADR о package boundary. `TokenSet`, `Resolve`, `CustomRegexpDetector`, observer/policy и `cmd/llmguard-eval` не меняют сигнатуры. Новые runtime dependencies не добавляются.
