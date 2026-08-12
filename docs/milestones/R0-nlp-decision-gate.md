# R0 — NLP decision gate и reference harness

| Поле | Значение |
|---|---|
| Change | `r0-nlp-decision-gate` |
| Capability | `nlp-reference` |
| Зависимости | M0 |
| Default variant | C |
| Результат | Проверенное решение по PERSON/ADDRESS runtime и воспроизводимый baseline |

## Outcome

Устранить главную архитектурную неопределённость NLP workstream до product
implementation. Сессия должна закончиться принятым ADR, точным dependency/license
audit и development-only reference harness. Это не порт Natasha/Yargy.

Исследование и архитектурное решение принадлежат Sol. После decision gate Composer
реализует только bounded tooling/harness и механически проверяемые fixtures.

## Scope

- Зафиксировать reference versions/commits Natasha `<1`, совместимого Yargy и
  относящихся к ним grammar/data sources.
- Построить transitive graph `NamesExtractor` и `AddrExtractor` до tokenizer,
  morphology, predicates, parser, interpretation и dictionaries.
- Извлечь фактически используемые constructs и grammemes.
- Проверить pure-Go tokenizer/morphology candidates, footprint и thread safety.
- Проверить лицензии кода, словарей, OpenCorpora-derived data и attribution.
- Выбрать direct Go grammars либо minimal generic rule runtime с rationale.
- Определить quality baseline и допустимые intentional differences.
- Добавить optional development reference harness и versioned JSONL schema; Python
  не входит в production/runtime dependencies и обычный library import path.

## Out of scope

- Реализация PersonDetector или AddressDetector.
- Полный порт Natasha, Yargy, Razdel или pymorphy.
- ML NER, ONNX, внешние API, автоматическая загрузка data в runtime.
- Требование 100% differential parity.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать upstream versions и собрать import/feature graph.
- [ ] Составить inventory grammemes, dictionaries и grammar primitives.
- [ ] Исследовать pure-Go tokenizer/morphology варианты и licenses.
- [ ] Сформировать decision matrix direct grammar vs generic runtime.
- [ ] Принять ADR по tokenizer, morphology source и parser strategy.
- [ ] Описать corpus schema и per-entity quality metrics.
- [ ] Реализовать version-pinned optional reference harness и sample fixtures.
- [ ] Проверить, что production Go module не получил Python/runtime dependency.

## Acceptance

1. `docs/natasha-port-scope.md` не содержит `?` в required-feature matrix и
   указывает точные reference revisions.
2. ADR однозначно выбирает architecture для M4/M5 и описывает rejected option.
3. Для каждого включаемого code/data source известны license, attribution и
   distribution strategy; неизвестная лицензия блокирует последующий milestone.
4. Reference harness воспроизводимо генерирует или проверяет versioned fixtures и
   не участвует в production runtime.
5. M4 и M5 можно поставить Composer без нового исследования базовой архитектуры.

## Verification

Focused команды уточняются после выбора harness, но обязательно включают:

```bash
go list -deps ./...
go test ./...
```

Если добавлен Python reference tool, его проверка выполняется отдельно от Go
runtime и с pinned environment. Milestone boundary:

```bash
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

## Exit evidence

- Archived change и canonical `openspec/specs/nlp-reference/spec.md`.
- `docs/natasha-port-scope.md`, ADRs, license inventory и fixtures/harness.
- Dashboard фиксирует выбранную strategy, не только статус `archived`.
