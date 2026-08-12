## Why

Перед реализацией RU PERSON и ADDRESS необходимо снять архитектурную неопределённость вокруг Natasha/Yargy: определить минимально достаточный pure-Go runtime, допустимые различия с reference implementation и лицензионно безопасные источники morphology/data. Без этого M4 и M5 рискуют либо преждевременно породить отдельный публичный порт, либо втянуть в `llm-guard` непропорционально большой generic parser и неаудированные словари.

## What Changes

- Переносится и уточняется workstream minimal Natasha-compatible subset из `docs/light_llm_guard_go_mvp_plan.md`: Natasha становится version-pinned development reference, а не целью API-совместимого порта.
- Вводится воспроизводимый dependency/feature/license audit для `NamesExtractor`, `AddrExtractor`, Yargy, tokenizer, morphology и используемых code/data sources.
- Фиксируется выбранная граница: `github.com/muonsoft/go-razdel` остаётся внешней зависимостью токенизации, а product-specific PERSON/ADDRESS runtime размещается внутри `llm-guard`; отдельный `go-natasha` и полный порт Yargy не создаются в рамках MVP.
- Фиксируется parser strategy: прямые Go-правила и bounded token matcher вместо generic Earley/Yargy-compatible runtime, если audit не обнаружит обязательного неподдерживаемого grammar construct.
- Определяются минимальный внутренний morphology contract, правила лицензирования/дистрибуции словарей и blocking gate для источников с неизвестным происхождением.
- Добавляются versioned JSONL schema, optional Python reference harness и sample fixtures для differential comparison без Python-зависимости production Go module.
- Определяются quality metrics и intentional differences для последующих M4 PERSON и M5 ADDRESS, включая консервативный false-positive profile и композиционное принятие адресов.

## Capabilities

### New Capabilities

- `nlp-reference`: воспроизводимый Natasha reference baseline, dependency/license decision gate и архитектурные ограничения минимального pure-Go NLP runtime для M4/M5.

### Modified Capabilities

- Нет.

## Impact

- Планируемые артефакты: `docs/natasha-port-scope.md`, ADR по tokenizer/morphology/parser strategy, license inventory, versioned fixtures и `tools/natasha-reference/`.
- Production-код PERSON/ADDRESS в этом change не создаётся; его реализация остаётся в M4 и M5.
- Будущая production-зависимость токенизации: `github.com/muonsoft/go-razdel`; Python, Natasha, Yargy и Pymorphy остаются development-only reference dependencies.
- Публичный API `llm-guard` не расширяется Natasha/Yargy-compatible DSL или extractor types.
