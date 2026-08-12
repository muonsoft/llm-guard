## Context

См. `proposal.md` — Why и `specs/safe-observability/spec.md`. Existing root package уже имеет immutable `Guard`, concurrent detector fan-out, deterministic `Resolve`, policy block до entropy, safe public errors и redacted `MaskResult`/`TokenSet`. Observability отсутствует; PERSON/ADDRESS/secrets имеют отдельные corpus harnesses с частично дублирующейся metric logic, а общего runner для всех 16 MVP entities и benchmark baseline нет.

M7 пересекает каждый public operation, security surfaces, testdata, CLI tooling и documentation. Core остаётся pure Go embeddable library: никакой exporter, goroutine worker, storage, proxy либо framework dependency не вводится.

## Goals / Non-Goals

**Goals:**

- определить минимальный immutable observer API с safe structured events и deterministic low-cardinality values;
- сделать success/error/block/restore-miss lifecycle наблюдаемым без изменения existing results/errors;
- отделить production-safe event path от conspicuous unsafe development diagnostics на уровне типов и options;
- заменить разрозненное измерение общим reusable internal evaluator и воспроизводимой CLI/report boundary;
- завершить M7 evidence: leakage matrix, full-entity quality report, benchmarks и executable README flow.

**Non-Goals:**

- async buffering, retries, observer error channel, persistent audit storage либо delivery guarantees;
- Prometheus/OpenTelemetry adapters, exporter server, logging integration и distributed tracing;
- production performance SLO или автоматическое сравнение benchmark numbers между разным hardware;
- изменение detector families, matching quality boundary, policy precedence, token format или public result/error semantics;
- публикация release/tag — это boundary M8.

## Decisions

### 1. Один safe terminal Event и один Observer method

Выбор: добавить string-like `Operation` (`detect`, `mask`, `restore`), `Outcome` (`success`, `error`, `blocked`, `restore_miss`), value-only `Event`, `Observer` с одним `Observe(Event)` method, `ObserverFunc` adapter, `NoopObserver` и `WithObserver`. Event содержит только integer/duration scalar fields и deterministic copied count slices; observer хранится в immutable `Guard`.

Rationale: три method-specific event types дублируют schema и усложняют generic metrics adapter. Один discriminated event остаётся framework-neutral и позволяет additive fields без exporter dependency. Альтернатива с map labels отвергнута: arbitrary keys/values создают high-cardinality и leakage surface.

Observer callback вызывается синхронно после terminal outcome; callback обязан быть concurrency-safe. Library не создаёт background queue и не владеет его lifecycle. Panic callback является panic caller-owned code и не преобразуется в Guard error, чтобы не скрывать дефект integration.

### 2. Instrumentation строится вокруг внутренних operation helpers

Выбор: public Detect/Mask/Restore ставят start time и defer terminal publication, а existing logic переносится в минимальные internal helpers там, где это нужно для единственного terminal event. Public `Mask` продолжает вызывать instrumented Detect и поэтому даёт Detect event плюс Mask event; каждый public call публикует ровно один event своего operation.

Rationale: defer обеспечивает coverage всех return paths без переписывания algorithm semantics. Отдельное подавление nested Detect event сделало бы lifecycle неожиданным и потребовало context flags. Event publication идёт после фиксации result/error и не участвует в их вычислении.

### 3. Counts представлены typed sorted slices

Выбор: `EntityCount` и `ActionCount` value types передаются в срезах, сортированных соответственно по entity/action. Built-in entity сохраняют canonical labels, а любые caller-defined entities агрегируются в один bounded `CUSTOM` label. Event включает input/output bytes, duration, total findings и restore misses. Mask action counts вычисляются по resolved findings; block event сохраняет zero output bytes. Detector names, spans и error categories кроме bounded outcome не включаются.

Rationale: slices стабильно marshal-ятся и не дают observer ссылку на internal maps. Map iteration nondeterministic, а labels по detector/error text потенциально высококардинальны и могут содержать caller-controlled data.

### 4. Restore miss определяется только для синтаксиса собственных placeholders

Выбор: Restore до replacement сканирует bounded token grammar `{{LLMG_<32 lowercase hex>_<4+ digits>}}` и считает каждое отсутствующее в caller TokenSet вхождение. Это исключает повторную интерпретацию восстановленного caller text как placeholder и сохраняет существующее поведение — неизвестные/изменённые tokens остаются unchanged. Обычный похожий текст не считается miss.

Rationale: generic brace substrings породили бы шум. Namespace/token никогда не попадают в event. Scanner переиспользуется только для counting и replacement semantics не меняет.

### 5. Unsafe development path отделён отдельными conspicuous symbols

Выбор: `UnsafeDevelopmentObserver`, `UnsafeDevelopmentEvent` и `WithUnsafeDevelopmentObserver` существуют независимо от safe observer. Raw event содержит operation/outcome, original input, operation output и copied findings; TokenSet mappings не экспортируются. GoDoc и README рядом с каждым entrypoint содержат `UNSAFE FOR PRODUCTION` warning.

Rationale: boolean `WithDiagnostics(true)` слишком легко включить через generic config и не показывает риск в code review. Отдельные имена дают явный grep-able opt-in. Дополнительный acknowledgement token сочтён ceremony без реальной security boundary; conspicuous compile-time API плюс документация достаточны.

### 6. Evaluation implementation остаётся internal, CLI — thin cmd entrypoint

Выбор: общий schema/loader/matcher/report formatter живёт в `internal/evaluation`; `cmd/llmguard-eval` только разбирает `-corpus`, `-format` и regression flags, собирает полный built-in Guard и вызывает runner. Unified `testdata/evaluation/cases.jsonl` хранит `schema_version`, stable id/category, input и expected `{entity,start,end}`; values synthetic. Exact match key — entity/start/end после Resolve.

Rationale: M7 требует reproducible repository runner, но не новый supported public package API до M8 review. Internal package устраняет duplicated corpus math и доступен CLI/tests. Отдельный process также позволяет генерировать ephemeral Markdown/JSON без core side effects.

Per entity TP/FP/FN считаются по exact spans. Negative-case denominator — число cases без expected span этой entity; false-positive case — такой case хотя бы с одним predicted span. `FPR = false_positive_cases / negative_cases`; `FNR = FN / (TP+FN)`. Нулевой denominator даёт `0`, явно документированный в report. Rows включают все 16 built-in entities даже при zero counts и сортируются по canonical entity order.

### 7. Один committed full-MVP corpus дополняет, а не заменяет family corpora

Выбор: unified corpus содержит минимум positive и negative opportunity для каждой entity плюс representative mixed RU и synthetic-secret cases. Existing detailed PERSON/ADDRESS/secrets corpora остаются authoritative family regression suites; evaluator не пытается объединять их несовместимые schemas автоматически.

Rationale: mechanical migration старых fixtures увеличила бы scope и риск. Небольшой common corpus доказывает общую metric/report boundary, а family suites продолжают проверять deep malformed/differential semantics.

### 8. Benchmarks отделяют Detect, Mask и Restore

Выбор: standard Go benchmarks имеют stable sub-benchmark names для representative RU, mixed и synthetic-secret inputs; setup/TokenSet создаются вне timed loop, Restore использует заранее masked result. `docs/benchmark-baseline.md` фиксирует environment и `-count=5` output; `docs/evaluation-baseline.md` генерируется exact CLI command.

Rationale: end-to-end-only benchmark скрывает источник allocation/regression. Отдельные operations дают actionable baseline; fixed synthetic inputs и offline execution делают его воспроизводимым. Hardware variability документируется, поэтому numbers не являются SLO.

### 9. Leakage audit использует уникальные synthetic markers

Выбор: tests форматируют safe events, MaskResult, TokenSet и errors через `%v`, `%+v`, `%#v`, JSON и observer capture для каждого outcome; затем проверяют отсутствие marker substrings. Unsafe observer проверяется отдельно и обязан раскрывать marker только после explicit option.

Rationale: field-by-field assertions не ловят accidental String/Marshal behavior. Marker tests дают end-to-end evidence без live PII/credentials и остаются устойчивы к безопасным additive fields.

## Risks / Trade-offs

- [Synchronous observer добавляет latency] → Noop default, один callback на terminal operation, никаких allocations text copies в safe path сверх bounded counts; benchmarks измеряют overhead.
- [Observer implementation data-races] → concurrency-safety явно входит в interface contract и проверяется race-safe capture fixture; core не сериализует caller callback глобальным lock.
- [Unsafe diagnostics попадут в production] → отдельные `Unsafe*` symbols, prominent GoDoc/README warning и отсутствие transitive activation из safe option.
- [Restore token scanner расходится с replacement grammar] → одна internal grammar/helper boundary и regression cases для unknown, mutated, cross-set и repeated placeholders.
- [Unified corpus создаёт ложное ощущение полного quality coverage] → docs называют его representative release-candidate baseline и ссылаются на более глубокие family corpora/known limitations.
- [Benchmark numbers нестабильны между hosts] → фиксировать environment, raw `-count=5` rows и no-SLO caveat; acceptance требует воспроизводимых names/command, не numeric gate.

## Migration Plan

Изменение additive: Guard без observability options сохраняет Noop behavior и existing operation semantics. Callers могут добавить safe `WithObserver`; unsafe path требует отдельного source change с conspicuous option. Rollback состоит в удалении options и новых tooling/docs; persistence, schema migration и external deployment отсутствуют.
