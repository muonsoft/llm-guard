# prefilter-evaluation Specification

## Purpose

Capability определяет воспроизводимую оценку llm-guard как precision-oriented prefilter: строгие гарантии на документированном контракте, диагностическое измерение незакрытой PII и проверку полного masking lifecycle без заявления high-recall/DLP-полноты.

## Requirements

### Requirement: Evaluation разделяет contract и exposure profiles
Evaluation system MUST предоставлять независимые `contract` и `exposure` profiles и MUST NOT объединять их результаты в единый release score. `contract` profile SHALL оценивать только документированные supported forms, а `exposure` profile MUST учитывать все размеченные sensitive spans, включая формы и source labels вне текущего product contract.

#### Scenario: Unsupported форма присутствует во внешнем corpus
- **WHEN** source annotation обозначена как `unsupported` для текущего detector contract
- **THEN** annotation исключается из contract TP/FN, но участвует в exposure coverage и leaked-sensitive-byte metrics

#### Scenario: Отчёты разных profiles публикуются вместе
- **WHEN** benchmark формирует contract и exposure results для одной suite revision
- **THEN** каждый результат имеет отдельные denominators, status и scope, а aggregate value не выдаётся за общий показатель промышленной полноты

### Requirement: Существующий offline conformance gate сохраняет строгую совместимость
Committed schema v1 corpus и его documented command MUST сохранять exact `(entity, start, end)` UTF-8 byte-span matching, полное positive/negative coverage всех MVP entities и non-zero exit при любом FP, FN или неполном coverage. Обычная установка, `go test ./...` и обязательный pull-request gate MUST NOT требовать network, external dataset cache, Python или external service.

#### Scenario: Известный supported case регрессирует
- **WHEN** schema v1 run получает хотя бы один FP, FN либо incomplete entity coverage
- **THEN** conformance command завершается non-zero и блокирует release readiness

#### Scenario: External cache отсутствует
- **WHEN** выполняются стандартные offline tests и conformance command в чистом checkout
- **THEN** проверки завершаются без попытки скачать external data

### Requirement: External и generated suites имеют versioned provenance
Каждая normalized suite MUST фиксировать schema version, suite ID, source record ID, source label, mapping policy version, UTF-8 input identity, exact source spans и disposition `supported`, `unsupported` либо `ignored`. Для `ignored` MUST быть непустая machine-readable reason. Каждый external source manifest MUST фиксировать canonical source, immutable revision или dated snapshot, cryptographic digest, license, attribution и distribution strategy; checksum, schema или provenance mismatch MUST отклонять run до evaluation.

#### Scenario: Pinned source успешно нормализован
- **WHEN** adapter читает source revision с совпадающим digest и поддерживаемой schema
- **THEN** он детерминированно создаёт normalized cases с исходными labels, byte spans, dispositions и source record IDs

#### Scenario: Source изменился без обновления manifest
- **WHEN** загруженный artifact не совпадает с manifest digest либо source schema больше не поддерживается
- **THEN** external run завершается non-zero до вызова detectors и не перезаписывает approved baseline

#### Scenario: Annotation исключается из оценки
- **WHEN** mapping policy присваивает source annotation disposition `ignored`
- **THEN** normalized record содержит явную reason, а отчёт отдельно показывает количество ignored annotations по source label и reason

### Requirement: Mapping policy сохраняет различие source truth и product contract
Mapping policy MUST хранить source label независимо от mapped llm-guard entity и MUST быть versioned отдельно от source snapshot. PERSON components MAY объединяться в contract annotation только по документированной композиционной политике; ADDRESS components MAY объединяться в contract annotation только при достаточной композиции для текущего ADDRESS detector. Неизвестные текущему продукту PII classes MUST NOT автоматически становиться clean negatives.

#### Scenario: Одиночное имя размечено как PII
- **WHEN** source содержит один FIRST_NAME без достаточной PERSON-композиции
- **THEN** annotation остаётся sensitive для exposure profile и обозначается unsupported для PERSON contract

#### Scenario: Адрес содержит только населённый пункт
- **WHEN** source содержит CITY без требуемой композиции STREET и HOUSE
- **THEN** annotation не считается contract ADDRESS, но остаётся sensitive в exposure profile

#### Scenario: Source label не имеет MVP entity
- **WHEN** source содержит OMS, DRIVER_LICENSE или другой unmapped sensitive label
- **THEN** label сохраняется в normalized data и exposure report и не используется как отрицательный пример для MVP entities

### Requirement: Contract report остаётся точным и per-entity
Contract report MUST вычислять TP, FP, FN, negative-case counts, false-positive-case counts, precision, recall, F1, FPR и FNR отдельно для каждой включённой entity с опубликованными zero-denominator rules. Match MUST требовать совпадение mapped entity и exact UTF-8 byte span; overlap и containment MAY выводиться только как diagnostics и MUST NOT повышать contract TP.

#### Scenario: Prediction пересекает gold span частично
- **WHEN** predicted span пересекает supported gold span, но его boundaries отличаются
- **THEN** contract report учитывает exact-span FP и FN, а возможный overlap diagnostic не изменяет gate status

#### Scenario: Suite покрывает подмножество entities
- **WHEN** external contract suite не содержит все MVP entities
- **THEN** report оценивает только объявленный suite scope и не применяет к нему schema v1 all-entity coverage rule

### Requirement: Exposure report измеряет оставшуюся чувствительную поверхность
Exposure report MUST вычислять по source label и mapped entity как минимум sensitive span count, fully covered span count, sensitive byte count, covered sensitive byte count, leaked sensitive byte count и overmatched byte count. Sensitive byte metrics MUST использовать объединение byte intervals, чтобы overlapping annotations или predictions не учитывались повторно. Отчёт MUST отдельно показывать supported и unsupported dispositions и MUST NOT интерпретировать exposure result как обещание high-recall detection.

#### Scenario: Prediction закрывает часть sensitive span
- **WHEN** resolved prediction покрывает только часть размеченного sensitive interval
- **THEN** span не считается fully covered, покрытые байты уменьшают leaked byte count только на фактически закрытую часть

#### Scenario: Prediction выходит за sensitive span
- **WHEN** resolved prediction включает байты вне объединения gold sensitive intervals
- **THEN** лишние байты учитываются как overmatched bytes без раскрытия их содержимого в report

### Requirement: Lifecycle profile проверяет фактическую защиту и обратимость
Lifecycle profile MUST выполнять configured `Detect → Resolve → policy → Mask → Restore` flow на synthetic либо legally redistributable cases. Для mask action restored text MUST побайтово совпадать с original input; для block action outbound text MUST отсутствовать; placeholder collision, mutation и miss MUST проверяться как отдельные outcomes. Lifecycle failure MUST завершать обязательный offline gate non-zero.

#### Scenario: Маскирование поддерживаемой PII обратимо
- **WHEN** supported case получает mask action и ответ содержит неизменённые placeholders
- **THEN** masked output не содержит защищаемые original spans, а Restore возвращает исходный UTF-8 input byte-for-byte

#### Scenario: Secret блокируется policy
- **WHEN** supported secret finding получает block action
- **THEN** lifecycle report фиксирует block без outbound text и без secret value в diagnostics

#### Scenario: Placeholder повреждён ответом
- **WHEN** synthetic response изменяет либо удаляет placeholder
- **THEN** lifecycle report фиксирует ожидаемый mutation или miss outcome и не считает round trip успешным

### Requirement: Evaluation artifacts не раскрывают sensitive values
Manifests, normalized metadata, reports и errors MUST NOT содержать raw credentials, действующие secrets, TokenSet mappings либо source substrings, кроме явно committed synthetic fixtures, прошедших provenance review. Reports SHALL идентифицировать failure через suite ID, source record ID, labels, spans, counts и hashes; добавление corpus или fixture MUST проходить license/provenance validation.

#### Scenario: External case даёт false negative
- **WHEN** evaluator формирует failure diagnostics для external source record
- **THEN** diagnostics содержат record ID, label и boundaries, но не печатают исходный sensitive substring или полный input

#### Scenario: Источник не прошёл license review
- **WHEN** manifest не содержит подтверждённую license, attribution либо разрешённую distribution strategy
- **THEN** source не включается в committed или release evaluation suite

### Requirement: Gates и thresholds являются явными и воспроизводимыми
Repository MUST разделять offline PR gate, explicitly invoked external benchmark и release evidence. Любой scoring threshold MUST быть versioned, иметь scope по profile/entity/source и содержать absolute boundary либо допустимый regression delta относительно именованного baseline. До утверждения initial audited baseline external exposure status MUST оставаться diagnostic; после утверждения изменение threshold или mapping policy MUST требовать нового baseline evidence.

#### Scenario: Pull request не меняет evaluation assets
- **WHEN** запускается обязательный PR workflow
- **THEN** он выполняет offline conformance и deterministic fixed-seed lifecycle/generated smoke без network dependency

#### Scenario: External benchmark запускается для release commit
- **WHEN** maintainer явно запускает external evaluation с доступным cache либо download permission
- **THEN** report фиксирует git commit, Go environment, suite/source revisions, mapping version, thresholds version и deterministic command

#### Scenario: Initial exposure baseline ещё не утверждён
- **WHEN** external suite впервые оценивается без approved threshold set
- **THEN** run сообщает diagnostic status и metrics, но не изображает произвольный score как release SLO
