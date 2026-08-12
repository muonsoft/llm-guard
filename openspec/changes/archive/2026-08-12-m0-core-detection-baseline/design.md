## Context

Repository пока содержит только module metadata и planning documents. Контракт поведения определён в `specs/core-detection/spec.md`; исходный draft API и выбор byte offsets находятся в `docs/light_llm_guard_go_mvp_plan.md`. Core должен стать самостоятельной корневой library package без provider-specific слоёв и без хранения PII.

## Goals / Non-Goals

**Goals:**

- Дать внешнему коду небольшой стабильный API `llmguard` для custom detectors.
- Сделать immutable `Guard`, который параллельно выполняет detectors и детерминированно собирает валидный результат.
- Сохранить программно проверяемые причины ошибок, не раскрывая их потенциально чувствительный текст через публичное `Error()`.
- Зафиксировать решения о UTF-8 byte offsets и stateless core как ADR и проверять baseline в CI.

**Non-Goals:**

- Built-in detectors, overlap resolution, masking, placeholders и restore.
- Recover от panic в пользовательском detector или принудительное завершение detector, игнорирующего context.
- Provider adapters, proxy, logging, audit, metrics, persistence и package hierarchy сверх корневого package.
- Гарантии binary/API compatibility до последующих MVP milestones.

## Decisions

### 1. Корневой package и закрытая модель options

Product API будет находиться в корне модуля с package name `llmguard`. `Detector` имеет только `Name() string` и `Detect(context.Context, string) ([]Finding, error)`. `Option` экспортируется как интерфейс с unexported apply-методом; `WithDetector` — единственный option этого milestone. Это оставляет расширение под контролем package и не усложняет detector contract.

`New` применяет options к локальному config, проверяет nil (включая typed nil), один раз считывает и валидирует имена, запрещает дубликаты и копирует entries в `Guard`. После возврата `Guard` конфигурация не меняется. Альтернатива с runtime registration отклонена: она потребовала бы locks и сделала бы результаты зависимыми от момента вызова.

### 2. Finding остаётся value-only metadata

`Finding` содержит `Entity`, `Start`, `End`, `Confidence`, `Detector`; поля исходного значения нет. Built-in entity constants переносятся из draft плана, но `EntityType` остаётся строковым для custom entities. Публичные комментарии и ADR явно определяют `[Start, End)` как UTF-8 byte offsets.

Validation выполняется после detector call и до агрегации: валидный UTF-8 input; непустая entity; finite confidence `[0,1]`; непустой in-range span; `utf8.RuneStart` для обеих границ (с `End == len(text)` как допустимой границей). Detector metadata может быть пустой и тогда заполняется сохранённым именем; любое другое имя отклоняется. Альтернатива доверять findings отклонена, поскольку downstream masking будет зависеть от корректных spans.

### 3. Один goroutine на detector и индексированные slots

`Detect` сначала проверяет caller context и input, затем создаёт производный context и запускает по одному goroutine на immutable detector entry. Каждый goroutine пишет только в свой заранее выделенный slot; `sync.WaitGroup` образует границу завершения. Первая ошибка инициирует cancel производного context, но метод ждёт все запущенные calls, поэтому после возврата нет принадлежащих core goroutines.

После ожидания caller context error имеет приоритет. Иначе выбирается substantive detector/validation error с минимальным registration index; ошибки отмены, возникшие только из-за sibling failure, не вытесняют исходный failure. При любой ошибке findings не возвращаются. Альтернатива вернуть partial findings отклонена: caller мог бы не заметить неполную защиту.

### 4. Стабильная сортировка с полным tie-break

Успешные validated findings получают внутренние registration и local indexes, объединяются и stable-sort по контрактному ключу `Start`, `End`, entity, detector, confidence descending, registration index, local index. В публичный результат копируются только `Finding`. Так scheduler timing никогда не влияет на порядок, а дубликаты одного detector сохраняют исходный порядок. Overlap resolution сознательно оставлен следующему change.

### 5. Безопасные типизированные ошибки

Package экспортирует sentinels для invalid config, invalid text, invalid finding и detector failure. Ошибки создаются/оборачиваются через `github.com/muonsoft/errors` согласно house convention. Для detector failure используется тип с безопасным `Error()` (operation и detector name) и отдельным `Unwrap()` к исходной причине: `errors.Is/As` сохраняются, но `%v` не печатает потенциально чувствительный текст причины. Finding validation сообщает detector, entity и категорию нарушенного поля, не исходный input/substring.

Альтернатива обычному `%w` в форматированной строке отклонена, потому что `Error()` такой цепочки раскрывает текст detector error. Полное удаление cause отклонено, поскольку ломает программную обработку caller errors.

### 6. Документация и repository baseline

Public API получает package documentation и black-box example с custom detector. Два ADR фиксируют byte offsets и stateless/caller-owned future masking state. GitHub Actions запускает `go test ./...`, `go vet ./...` и `go test -race ./...` на поддерживаемой версией `go.mod` toolchain. Focused unit tests остаются stdlib-first; test-only helper dependency допустима только если реально сокращает проверки.

## Risks / Trade-offs

- [Detector игнорирует context и блокируется] → Core не может безопасно убить goroutine; контракт документирует cooperative cancellation, а `Detect` ждёт detector для отсутствия утечек после возврата.
- [Один goroutine на detector не ограничивает fan-out] → В M0 число detectors задаёт caller; bounded worker pool можно добавить отдельным option при доказанной необходимости.
- [Имя detector или custom error может само содержать sensitive text] → Имя обязано быть непустым и не содержать управляющих символов; public documentation запрещает помещать sensitive data в metadata. Исходный error остаётся доступным только программно через unwrap.
- [Стабильный sort отличается от будущего overlap priority] → Resolver следующего milestone получает уже детерминированный input и может определить отдельный output order без изменения raw detection safety.
- [Новая dependency на error package] → Ограничить runtime dependencies `github.com/muonsoft/errors`; CI и `go mod tidy` фиксируют полный module graph.

## Migration Plan

Change добавляет первый product API, поэтому миграция существующих callers не нужна. Rollback состоит в revert milestone commit до появления downstream milestone; persisted state и data migration отсутствуют.
