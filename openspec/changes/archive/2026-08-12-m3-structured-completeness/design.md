## Context

M0–M2 уже закрепили string-based `EntityType`, concurrent `Detector`, immutable `Guard`, UTF-8 byte spans, deterministic resolver и caller-owned reversible token state. См. `proposal.md` и delta specs этого change. Три оставшиеся structured entity требуют context-first matching, а regexp adapter должен использовать те же validation/error boundaries без отдельного pipeline или runtime registration.

## Goals / Non-Goals

**Goals:**

- Закрыть structured PII MVP набор консервативными pure-Go detectors и общим corpus contract.
- Добавить минимальный compile-once regexp adapter, безопасный для concurrent use.
- Сохранить существующие core, resolver и masking contracts без breaking changes.
- Зафиксировать после реализации public API review в `docs/adr/0004-mvp-public-api-boundary.md`.

**Non-Goals:**

- Внешняя проверка паспортов/счетов, BIK lookup и утверждение юридической валидности реквизитов.
- Runtime detector registry, plugin ABI, DSL, named capture semantics или overlapping regexp enumeration.
- Расширение regexp adapter контекстными правилами: custom boundaries остаются частью caller pattern.

## Decisions

### 1. Один candidate-normalize-validate pattern для built-in detectors

Каждый detector остаётся отдельным stateless типом с публичным constructor, как detectors M1/M2. Regex выделяет ограниченный candidate, затем helper проверяет boundary, разрешённую форму, context и доступную semantic validation; Finding создаётся только после всех проверок. Это сохраняет простое reviewable поведение и не вводит общий DSL, который для трёх разных контекстов был бы сложнее прямого кода.

Контекст ищется в исходном тексте только в пределах той же строки. Между ближайшим предшествующим marker и candidate допускается не более 48 UTF-8 bytes, причём gap может состоять только из whitespace, `:`, `№`, `#`, `-`, `,` и коротких слов формы (`РФ`, `серия`, `номер`). Marker comparison case-insensitive; найденный span никогда не расширяется на marker. Это ограничивает ложные связи между реквизитами соседних строк. Альтернатива — произвольное paragraph window — отклонена как слишком permissive.

### 2. Паспорт поддерживает только contiguous numeric representation

Поддерживаются `NNNN NNNNNN` и `NN NN NNNNNN` с одним типом ASCII space между группами. После normalization должно быть ровно десять цифр, но checksum не заявляется. Раздельная запись с `номер` посередине не объединяется в один finding: маскирование такого interval захватило бы несекретный marker, а два findings ошибочно представляли бы один паспорт двумя сущностями. Расширение потребует отдельного composition contract.

### 3. Банковский счёт использует context-first fallback и локальный БИК

Account candidate — ровно 20 ASCII digits compact либо пять четырёхзначных групп с одиночными ASCII spaces. Homogeneous digits и numeric extensions отклоняются. Marker обязателен. Если на той же строке найден отдельный девятизначный `БИК` рядом с marker, используется российская проверка: последние три цифры БИК конкатенируются с 20 цифрами счёта, циклические веса `7,1,3`, сумма произведений modulo 10 должна быть нулевой. При наличии БИК failed checksum запрещает finding; при отсутствии БИК форма остаётся явно документированным context-first match. Альтернатива «только checksum» отвергнута, потому что standalone account часто приходит без БИК; «любые 20 цифр» отвергнута из-за false positives.

### 4. Дата рождения отделена от generic date parsing

Detector поддерживает numeric `DD.MM.YYYY`, `DD/MM/YYYY` и русские названия двенадцати месяцев в родительном падеже. Calendar validity проверяется через `time.Date` с round-trip компонентов; год содержит ровно четыре цифры. Birth marker обязателен в bounded preceding context. Общий date detector не создаётся, поэтому договорные даты не попадают в PII. Альтернатива guessed date locale или двузначный год отклонена как неоднозначная.

### 5. Public regexp adapter компилируется один раз

Добавляются:

```go
type RegexDetectorConfig struct {
    Name       string
    Entity     EntityType
    Pattern    string
    Confidence float64
}

func NewCustomRegexpDetector(config RegexDetectorConfig) (*CustomRegexpDetector, error)
```

`CustomRegexpDetector` экспортируется с неэкспортируемыми immutable полями и реализует `Detector`; `regexp.Regexp` компилируется constructor-ом и безопасно переиспользуется конкурентно. Constructor использует `newInvalidConfigError` только с фиксированными reason strings и не wraps regexp syntax error, поскольку она может содержать caller pattern. Возвращать `Detector` interface прямо из constructor отклонено: concrete return сохраняет обычную Go discoverability и всё равно не раскрывает mutable state.

`Detect` использует `FindAllStringIndex`, пропускает `[start,start)` matches, проверяет context между matches и возвращает full-match byte spans. Go RE2 гарантирует корректные byte indexes для валидного UTF-8 input, а Guard повторно применяет общую finding validation. Неявные word boundaries не добавляются, чтобы adapter не менял caller regexp semantics.

### 6. Resolver сохраняет built-in приоритет над custom entity

String entity, которой нет в `entityPriority`, остаётся на fallback priority `0`; все текущие built-ins имеют положительный priority. Это даёт deterministic overlap behavior без нового public priority API. Non-overlapping custom findings проходят существующие Mask/Restore без изменений. Configurable public priorities отложены до появления реального use case и отдельного compatibility review.

### 7. Verification объединяет unit, corpus, fuzz и API evidence

Focused tests покрывают constructors, supported/unsupported forms, unsafe errors, zero-width/full-match behavior, resolver overlap, custom Go/regexp round trips и concurrency. `FuzzStructuredDetectorsInvariants` расширяется новыми built-ins, а отдельный `FuzzCustomRegexpDetectorInvariants` проверяет UTF-8 spans, zero-width safety и deterministic output. Точные bounded команды:

```bash
go test ./... -run 'Test(Passport|BankAccount|DateOfBirth|Regex|Custom)'
go test . -run '^$' -fuzz '^FuzzStructuredDetectorsInvariants$' -fuzztime=2s
go test . -run '^$' -fuzz '^FuzzCustomRegexpDetectorInvariants$' -fuzztime=2s
```

После implementation orchestrator выполняет полный public API review, фиксирует accepted surface/extension rules в ADR 0004 и проверяет, что M4–M7 могут добавлять detectors/policy поверх существующих contracts без breaking core edits.

## Risks / Trade-offs

- [Context window может пропустить необычную верстку реквизитов] → формы и line/window limits документируются как supported subset; corpus содержит multiline negatives.
- [20-digit account без БИК не имеет сильной semantic validation] → marker обязателен, fallback явно не называется внешней валидацией, а доступный БИК всегда усиливает проверку.
- [Regexp с дорогим выражением расходует CPU] → Go RE2 исключает exponential backtracking; runtime budget/cancellation внутри одного regexp call не обещается.
- [Custom pattern может намеренно захватывать слишком широкий текст] → adapter соблюдает exact caller semantics, а safe error/finding validation предотвращает повреждённые spans, но не заменяет caller review.
- [Новые detectors конфликтуют с существующими numeric entities] → built-in priority и end-to-end overlap corpus проверяются детерминированно.

## Migration Plan

Изменение additive: существующий код и constructors остаются совместимыми. Новые users явно регистрируют detectors через `WithDetector`. Rollback состоит в удалении новых registrations; форматы TokenSet и existing entities не меняются.
