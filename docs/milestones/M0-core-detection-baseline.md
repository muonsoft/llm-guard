# M0 — Core detection baseline

| Поле | Значение |
|---|---|
| Change | `m0-core-detection-baseline` |
| Capability | `core-detection` |
| Зависимости | нет |
| Default variant | C |
| Результат | Полезная detection-only Go library с расширяемым Detector API |

## Outcome

Создать минимальный, конкурентно безопасный core, который принимает публичные
detectors, валидирует их findings и возвращает детерминированный detection result.
Это самостоятельная shipping boundary: пользователь уже может подключить свой
detector без masking и без привязки к LLM provider.

## Scope

- Компактный public API: `EntityType`, built-in entity constants, `Finding`,
  `Detector`, `Guard`, `New`, options и `Detect`.
- `Detector` сообщает стабильное имя и принимает `context.Context` и текст.
- UTF-8 byte offsets `[Start, End)` документированы и проверяются core.
- Validation custom findings: entity, bounds, UTF-8 boundaries, confidence и
  detector metadata без включения sensitive value.
- Детерминированная агрегация и порядок findings при нескольких detectors.
- Нет global mutable state; инициализированный `Guard` безопасен для concurrent use.
- Fake/custom detector tests, race coverage и repository CI baseline.
- ADR: byte offsets; stateless core/caller-owned future state.

## Out of scope

- Built-in PII detectors, overlap resolver, masking, restore и `TokenSet`.
- Audit/metrics API, secrets, NLP и provider adapters.
- Generic policy engine и сложная package hierarchy.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать public contracts и failure semantics в delta spec/design.
- [ ] Реализовать entity и finding types с безопасным представлением.
- [ ] Реализовать Detector contract, options и immutable Guard construction.
- [ ] Реализовать concurrent detection pipeline и context cancellation.
- [ ] Реализовать validation и deterministic aggregation.
- [ ] Добавить positive, invalid-span, Unicode, error и race tests.
- [ ] Добавить ADR и package documentation.
- [ ] Настроить минимальную CI-проверку, не вводя лишний framework.

Фактический `tasks.md`, созданный FF, является авторитетным; список выше задаёт
coverage и не требует отдельного Composer job на каждый пункт.

## Acceptance

1. Внешний Go-код может создать `Guard` с custom detector и вызвать `Detect`.
2. Findings имеют валидные UTF-8 byte ranges и стабильный порядок; невалидный
   detector output отклоняется безопасной ошибкой без sensitive substring.
3. `Guard` корректно обрабатывает context cancellation и detector errors согласно
   design, не скрывая частичный failure.
4. Один Guard проходит concurrent race test без global mutable state.
5. Core не импортирует provider, proxy, logging или persistence dependencies.

## Verification

Focused:

```bash
go test ./... -run 'TestGuard|TestFinding|TestDetector'
go test -race ./... -run 'TestGuard'
```

Milestone boundary:

```bash
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

## Exit evidence

- Archived change и canonical `openspec/specs/core-detection/spec.md`.
- Public API example компилируется в test/example.
- Review подтверждает отсутствие sensitive values в errors/formatting.
- Dashboard и orchestration experiment обновлены.
