# M7 — Safe observability и MVP release candidate

| Поле | Значение |
|---|---|
| Change | `m7-safe-observability-rc` |
| Capabilities | `safe-observability`, при необходимости существующие core specs |
| Зависимости | M5 и M6 |
| Default variant | C |
| Результат | Полный MVP с production-safe defaults и измеримым quality profile |

## Outcome

Довести функционально полный guard до release-candidate состояния: безопасные
observer/audit contracts, opt-in unsafe diagnostics, metrics abstraction,
evaluation runner, benchmarks и рабочий embedded example.

## Scope

- Noop observer по умолчанию; framework-neutral detection/mask/restore events.
- Production/hard events: lengths, durations, counts by entity/action, misses — без
  original/masked text, findings values и TokenSet mappings.
- Development/soft diagnostics только через явный opt-in с conspicuous warning.
- Metrics/observer abstraction без Prometheus dependency в core.
- Stable event semantics при success, detector error, block и restore miss.
- Evaluation corpus runner с per-entity precision/recall/F1/FPR/FNR.
- Benchmarks и baseline report для representative RU/mixed prompts.
- README embedded flow, security considerations и known limitations.
- Full unsafe logging/formatting/error audit.

## Out of scope

- Обязательный logging/metrics framework, exporter server и dashboard.
- Production SLO guarantee, distributed tracing и persistent audit store.
- OpenAI adapters/proxy и новые detector families.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать hard/soft profiles и event schemas.
- [ ] Реализовать observer lifecycle и Noop default.
- [ ] Реализовать explicit unsafe development diagnostics.
- [ ] Добавить framework-neutral metrics event/counter integration points.
- [ ] Реализовать corpus evaluation runner и per-entity report.
- [ ] Добавить benchmarks и representative fixtures.
- [ ] Провести sensitive-data leakage audit и regression tests.
- [ ] Обновить README example, security notes и known limitations.

## Acceptance

1. Default observer/events/errors никогда не содержат original text, raw findings,
   TokenSet mappings или secret values.
2. Unsafe diagnostics невозможно включить неявно; API/docs явно маркируют риск.
3. Metrics/events различают entity, action, block и restore miss без high-cardinality
   sensitive labels.
4. Evaluation выдаёт отдельные метрики по всем MVP entities; benchmark baseline
   воспроизводим.
5. README example компилируется и демонстрирует полный embedded flow.

## Verification

```bash
go test ./... -run 'Test(Observer|Audit|Metrics|Redact|Evaluation|Example)'
go test ./... -run '^$' -bench . -benchmem
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Milestone packet должен назвать точную команду генерации quality report и место
baseline output; generated ephemeral reports не коммитятся без причины.

## Exit evidence

- Archived `safe-observability` spec.
- Full MVP evaluation/benchmark summary.
- Review checklist по всем sensitive-data surfaces.
- Версия готова к OSS stabilization, но ещё не объявляется release.
