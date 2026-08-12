## 1. Core

- [x] 1.1 Добавить public `Operation`, `Outcome`, `EntityCount`, `ActionCount` и safe `Event` value contract с JSON/formatting-safe fields.
- [x] 1.2 Добавить concurrency-safe `Observer` contract, `ObserverFunc`, `NoopObserver` и immutable `WithObserver` option с nil validation.
- [x] 1.3 Добавить отдельные `UnsafeDevelopmentObserver`, `UnsafeDevelopmentEvent` и `WithUnsafeDevelopmentObserver` с conspicuous GoDoc warning и nil validation.
- [x] 1.4 Расширить immutable `guardConfig`/`Guard` observer fields без mutable globals и сохранить default Noop behavior.

## 2. Detectors

- [x] 2.1 Инструментировать public `Guard.Detect` одним terminal safe event на success/error/cancellation, сохранив deterministic detector fan-out и error selection.
- [x] 2.2 Добавить deterministic sorted per-entity count builder без spans, detector names, raw/normalized values или error causes.
- [x] 2.3 Добавить Detect lifecycle tests для no-detector, success, detector error, invalid text, cancellation и nested Mask detection.

## 3. Masking

- [x] 3.1 Инструментировать `Guard.Mask` terminal events для success, ordinary error и policy block с zero block output и per-entity/per-action counts.
- [x] 3.2 Инструментировать `Guard.Restore` terminal events для success, ordinary error и restore miss без изменения existing unknown-token behavior.
- [x] 3.3 Реализовать bounded scanner общей placeholder grammar и miss counting для unknown, mutated, cross-set и repeated tokens без публикации namespace/token strings.
- [x] 3.4 Добавить Mask/Restore lifecycle tests для allow/mask/block, empty results, entropy error, nil TokenSet, invalid UTF-8 и restore miss.

## 4. Audit

- [x] 4.1 Реализовать safe callback dispatch с copied deterministic count slices и non-negative durations для concurrent calls одного Guard.
- [x] 4.2 Реализовать explicit unsafe development callbacks для Detect/Mask/Restore с raw text/findings только при отдельном `Unsafe*` option и без TokenSet mappings.
- [x] 4.3 Добавить in-memory metrics observer test, считающий requests, entity/action counts, blocks и restore misses только по safe event fields.
- [x] 4.4 Добавить marker-based leakage matrix для safe events, `%v`, `%+v`, `%#v`, JSON, errors, `MaskResult` и `TokenSet` на success/error/block/miss paths.
- [x] 4.5 Добавить concurrent observer/unsafe-observer regression tests, пригодные для `go test -race ./...`, без serialization callback core lock.
- [x] 4.6 Провести repository audit logging/formatting/error surfaces и зафиксировать checklist/result без raw fixture values.

## 5. Tests и evaluation

- [x] 5.1 Добавить `internal/evaluation` schema types и strict JSONL loader с validation schema version, ids, categories, UTF-8 и exact byte spans.
- [x] 5.2 Реализовать Detect → Resolve evaluator с exact `(entity,start,end)` TP/FP/FN, negative/false-positive cases и documented zero-denominator formulas.
- [x] 5.3 Реализовать deterministic Markdown и JSON report formatters со всеми 16 built-in entity rows и aggregate summary в canonical order.
- [x] 5.4 Добавить thin `cmd/llmguard-eval` с `-corpus`, `-format` и mandatory regression exit policy без network/external services.
- [x] 5.5 Добавить `testdata/evaluation/cases.jsonl` с synthetic positive/negative coverage всех MVP entities, mixed RU и synthetic-secret cases.
- [x] 5.6 Добавить evaluator unit tests для schema rejection, exact mismatch, FP/FN/FPR/FNR math, deterministic ordering/format и cancellation/error handling.
- [x] 5.7 Добавить CLI integration test и запустить `go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format markdown -fail-on-regression`.
- [x] 5.8 Добавить stable Detect/Mask/Restore benchmarks для representative RU, mixed PII и synthetic secret inputs с setup вне timed loops.
- [x] 5.9 Запустить focused suite `go test ./... -run 'Test(Observer|Audit|Metrics|Redact|Evaluation|Example)'`.
- [x] 5.10 Запустить benchmark suite `go test ./... -run '^$' -bench . -benchmem -count=5` и сохранить environment/output evidence.
- [x] 5.11 Запустить broad checks `go test ./...`, `go vet ./...`, `go test -race ./...`.

## 6. Docs

- [x] 6.1 Добавить generated `docs/evaluation-baseline.md` с exact CLI command, formulas, full per-entity report и links на detailed family corpora.
- [x] 6.2 Добавить `docs/benchmark-baseline.md` с Go/OS/arch/CPU, exact command, stable raw rows и explicit no-SLO caveat.
- [x] 6.3 Обновить README status и embedded flow для release candidate, safe observer/metrics integration и explicit unsafe development warning.
- [x] 6.4 Добавить README security considerations, known limitations, default secret block, caller-owned TokenSet и evaluation/benchmark reproduction commands.
- [x] 6.5 Обновить executable package example Detect → Mask → caller-owned LLM boundary → Restore и подтвердить `go test ./... -run 'TestExample'`/example execution.
- [x] 6.6 Проверить `go.mod`/imports на отсутствие Prometheus, OpenTelemetry, logging/exporter, network, storage или proxy dependencies.
- [x] 6.7 Выполнить `openspec validate m7-safe-observability-rc --strict --no-interactive` и подготовить requirement/scenario verification evidence.
