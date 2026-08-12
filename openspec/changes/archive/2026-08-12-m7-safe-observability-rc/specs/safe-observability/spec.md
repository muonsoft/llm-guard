## Purpose

Определяет безопасный по умолчанию observability contract и воспроизводимые quality/benchmark evidence для полного embeddable MVP без зависимости core от logging или metrics framework.

## ADDED Requirements

### Requirement: Safe observer является additive и выключен по умолчанию
Guard SHALL поддерживать immutable framework-neutral observer configuration для операций detect, mask и restore. Без явной configuration MUST использоваться Noop behavior без callbacks, side effects, exporter, network, logging либо metrics dependency; один настроенный Guard MUST оставаться безопасным для concurrent use.

#### Scenario: Guard без observer
- **WHEN** caller создаёт Guard без observability options и выполняет detect, mask и restore
- **THEN** результаты и ошибки совпадают с существующим operation contract, а observability не создаёт output или external side effects

#### Scenario: Concurrent observer calls
- **WHEN** один настроенный Guard конкурентно обрабатывает несколько операций
- **THEN** каждая операция публикует отдельный завершённый event без shared mutable event state или data race

### Requirement: Production events содержат только bounded safe fields
Production event schema MUST ограничиваться operation, stable outcome, input/output byte lengths, non-negative duration, counts по built-in/custom entity и action, общим finding count и restore miss count. Schema MUST NOT содержать original, masked или restored text, matched fragments, raw/normalized values, detector error cause/message, placeholder namespace, token strings либо TokenSet mappings; entity/action collections MUST иметь deterministic order и не создавать sensitive high-cardinality labels.

#### Scenario: Успешная mask операция
- **WHEN** mask успешно разрешает allow и mask findings
- **THEN** event сообщает safe lengths, duration и counts по entity/action без текста, spans, detector names или token material

#### Scenario: Ошибка detector
- **WHEN** detect либо вложенный detect в mask завершается detector error
- **THEN** соответствующий error outcome не содержит error text/cause или часть исходного input

#### Scenario: Formatting production event
- **WHEN** production event форматируется через `%v`, `%+v`, `%#v` или JSON
- **THEN** ни один запрещённый sensitive value не появляется в результате

### Requirement: Lifecycle и outcomes являются стабильными
Observer MUST получать один terminal event на каждый public Detect, Mask и Restore call, который дошёл до operation body; Mask MAY дополнительно породить terminal Detect event для своей вложенной detection phase. Stable outcomes SHALL различать success, error, policy block и restore miss; block и restore miss MUST быть доступны как low-cardinality machine-readable значения, а observer не MUST получать partial success event.

#### Scenario: Policy block
- **WHEN** resolved finding получает action block
- **THEN** Mask event имеет block outcome, zero output length, counts blocked entities/actions и не раскрывает blocked value

#### Scenario: Restore miss
- **WHEN** response содержит syntactically valid llm-guard placeholder, отсутствующий в переданном TokenSet
- **THEN** Restore сохраняет существующий unchanged-placeholder result и публикует restore-miss outcome с bounded miss count

#### Scenario: Обычная ошибка restore
- **WHEN** Restore получает invalid input или token set
- **THEN** event имеет error outcome без error message, response text или mapping material

### Requirement: Metrics integration не требует exporter framework
Safe observer event MUST предоставлять достаточные low-cardinality integration points для requests, durations, findings, actions, blocks и restore misses по operation/entity/action/outcome. Core module MUST NOT зависеть от Prometheus, OpenTelemetry, logging framework или exporter runtime.

#### Scenario: Caller считает metrics
- **WHEN** caller реализует observer как in-memory counters
- **THEN** он может отдельно посчитать detect/mask/restore requests, entity/action counts, blocks и restore misses только по safe event fields

### Requirement: Unsafe diagnostics требуют conspicuous explicit opt-in
Raw development diagnostics SHALL предоставляться только через отдельный public API, названия типов, option и documentation которого явно содержат `Unsafe` и предупреждение о production risk. Safe observer configuration MUST NOT автоматически активировать unsafe callbacks; unsafe event MAY включать original/masked/restored text и raw findings, но MUST NOT автоматически сериализовать TokenSet mappings.

#### Scenario: Настроен только safe observer
- **WHEN** caller передаёт production observer без unsafe option
- **THEN** raw development callback не вызывается и raw values недоступны через safe event

#### Scenario: Явный development opt-in
- **WHEN** caller явно настраивает unsafe development observer
- **THEN** он получает документированный raw diagnostic event и принимает обозначенный риск утечки sensitive data

### Requirement: Evaluation runner вычисляет per-entity quality profile
Repository SHALL содержать offline deterministic runner и versioned annotated corpus, покрывающий каждую MVP entity: PERSON, ADDRESS, EMAIL, PHONE, IP_ADDRESS, URL, INN, SNILS, PASSPORT, BANK_CARD, BANK_ACCOUNT, DATE_OF_BIRTH, SECRET_JWT, SECRET_PRIVATE_KEY, SECRET_API_KEY и CONNECTION_STRING. Report MUST выдавать отдельные TP, FP, FN, negative-case counts, precision, recall, F1, FPR и FNR для каждой entity, zero-denominator rules и aggregate summary; matching MUST использовать entity плюс exact UTF-8 byte span.

#### Scenario: Воспроизводимый full-corpus report
- **WHEN** developer запускает документированную CLI-команду на committed corpus
- **THEN** runner валидирует schema, выполняет Detect → Resolve и печатает deterministic rows для всех MVP entities в стабильном порядке

#### Scenario: Ошибочная prediction
- **WHEN** predicted entity/span отсутствует в expected annotations либо expected span не найден
- **THEN** соответствующие per-entity FP/FN и производные rates детерминированно отражают mismatch и команда возвращает non-zero при configured mandatory regression boundary

### Requirement: Benchmarks имеют воспроизводимый representative baseline
Repository MUST содержать benchmarks минимум для Detect, Mask и Restore на representative RU structured/person/address prompt, mixed PII prompt и synthetic secret prompt. Committed baseline report SHALL фиксировать Go version, OS/architecture, CPU, точную command, benchmark names, ns/op, B/op, allocs/op и caveat, что baseline не является production SLO.

#### Scenario: Benchmark reproduction
- **WHEN** developer выполняет документированную benchmark command
- **THEN** все benchmark cases запускаются без network/external services и output можно сопоставить с committed baseline по стабильным names

### Requirement: Release-candidate документация демонстрирует безопасный embedded flow
README и executable examples MUST показывать construction, Detect или Mask, caller-owned LLM boundary и Restore с safe observability default. Documentation SHALL явно описывать unsafe diagnostics, security considerations, errors/formatting boundaries, evaluation/benchmark commands и known MVP limitations; версия MUST называться release candidate, а не объявленным release.

#### Scenario: Embedded example компилируется
- **WHEN** выполняются package example tests
- **THEN** documented Detect → Mask → Restore flow компилируется и подтверждает byte-for-byte restore без external service

#### Scenario: Security review документации
- **WHEN** maintainer проверяет release-candidate README
- **THEN** он видит запрет production raw diagnostics, caller ownership TokenSet, default secret block и отсутствие SLO/exporter guarantees

### Requirement: Sensitive-data surface проходит regression audit
Errors, observer events, diagnostics disabled path, formatting and JSON surfaces MUST быть проверены synthetic marker-based tests. Default operations MUST NOT отправлять original/masked/restored text, finding values, detector causes, placeholder/token material или TokenSet mappings в observer и standard formatting; тесты MUST покрывать success, detector error, block, restore miss и concurrent calls.

#### Scenario: Marker leakage regression
- **WHEN** tests пропускают уникальные synthetic sensitive markers через все safe lifecycle paths и formatting forms
- **THEN** ни один marker не присутствует в production events, errors или serialized safe values
