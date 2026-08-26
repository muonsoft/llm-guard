## Context

См. `proposal.md` — Why. Сейчас корневой `package llmguard` содержит и публичный контракт ADR 0004 (`Guard`, `Detector`, `Finding`, constructors), и тела detectors. PERSON/ADDRESS уже следуют нужному паттерну: `internal/nlp` возвращает spans, корень мапит их в `Finding`. Structured PII и secrets всё ещё живут в корне и тянут пакетные хелперы (`detectutil.go`, `contextmatch.go`, `precedingRune` в `email.go`).

Ограничения: `internal/detect` не должен импортировать `llmguard` (иначе цикл с constructors). Fuzz smoke в `scripts/release-check.sh` запускает точные targets в пакете `.`. Внешний consumer проверяется только публичным module path.

## Goals / Non-Goals

**Goals:**

- Один пакет `internal/detect` для structured PII, secrets и shared text/RU helpers.
- Корневые constructors остаются source-compatible и мапят spans в `Finding` с прежними entity, detector name и confidence.
- Сохранить детерминизм, UTF-8 byte spans, context cancellation и concurrent-safe detectors.

**Non-Goals:**

- Публичные подпакеты `detect/`, `resolver/`, `masking/` из черновика MVP-плана.
- Type aliases публичных типов на `internal/core`.
- Перенос `Resolve`, `Mask`/`TokenSet`, `placeholder.go`, `validate.go` и `CustomRegexpDetector`.
- Смена fuzz target names или перенос `package llmguard_test` контрактных тестов.

## Decisions

### 1. Spans в internal, Finding только в корне

`internal/detect` экспортирует `Span{Start, End int}` и функции вида `Email(ctx, text) ([]Span, error)`. Confidence, `EntityType` и detector name задаёт корневой адаптер — как в текущих `person.go`/`address.go`.

Альтернатива «internal реализует `llmguard.Detector`» отвергнута из-за import cycle. Альтернатива «type alias Finding на internal/core» отвергнута: хуже godoc и лишний риск для ADR 0004.

### 2. Один пакет `internal/detect`, не пакет на сущность

~16 файлов и общие helpers живут вместе. Дробить `structured`/`secret` нет нужды: зависимости одинаковые, а мелкие packages запрещены планом MVP §19.

### 3. Один unexported `spanDetector` в корне

Все built-in constructors, включая PERSON/ADDRESS, возвращают один immutable тип с `detect func(context.Context, string) ([]detect.Span, error)`. NLP spans конвертируются в `detect.Span` тонкой обёрткой. Это убирает 16 почти одинаковых файлов без смены публичных имён constructors.

### 4. Whitebox fuzz остаётся в корневом пакете

`FuzzEmailDetectorBoundaries` продолжает жить в `fuzz_test.go` (пакет `.`). Для mailbox/boundary invariants он может вызывать exported helpers `internal/detect` — это внутри модуля и не расширяет public API. Имена fuzz targets не меняются.

### 5. ADR фиксирует layout, specs не трогаем

Поведение не меняется, поэтому `skip_specs: true`. Решение о границе пакетов записывается в `docs/adr/0005-internal-detect-layout.md`, не в product specs.

## Risks / Trade-offs

- [Случайный `import llmguard` из `internal/detect` создаст cycle] → пакет работает только со `Span` и стандартной библиотекой; constructors остаются в корне.
- [Расхождение confidence/detector name при переносе] → адаптер явно задаёт прежние константы; существующие `llmguard_test` и corpus остаются регрессией.
- [Корневой каталог всё ещё содержит много `*_test.go`] → это ограничение Go; контрактные тесты специально остаются рядом с public API.

## Migration Plan

Перенос внутри одного module: удалить старые корневые файлы detectors после появления адаптеров. Rollback — revert коммита. Для внешних consumer миграция не требуется.
