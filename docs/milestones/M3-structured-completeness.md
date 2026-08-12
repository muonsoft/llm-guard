# M3 — Structured completeness и custom regexp

| Поле | Значение |
|---|---|
| Change | `m3-structured-completeness` |
| Capabilities | `structured-pii`, `custom-detection` |
| Зависимости | M2 |
| Default variant | C |
| Результат | Закрытый structured PII scope и стабильная extensibility boundary |

## Outcome

Добавить консервативные PASSPORT, BANK_ACCOUNT и contextual DATE_OF_BIRTH,
публичный `CustomRegexpDetector` и провести первый полный API review перед NLP.

Custom Go detector уже должен поддерживаться core M0; здесь проверяется его
совместимость с resolver/masking и добавляется удобный безопасный regexp adapter.

## Scope

- Паспорт РФ только по надёжному формату/контексту; не любая серия цифр.
- Минимально определённые банковские счета РФ с normalization и validation,
  возможной только при имеющихся данных; context-first fallback документирован.
- DATE_OF_BIRTH только рядом с явными RU birth-context markers.
- Config/API `CustomRegexpDetector`: name, string entity, compiled pattern,
  confidence и безопасная construction error.
- Validation regexp findings и zero-width/boundary handling.
- Совместимость custom detectors с priorities, masking, restore и concurrency.
- Public API review/ADR для изменений M0–M3.
- Corpus для обычных дат, договорных реквизитов и конфликтующих numeric strings.

## Out of scope

- Generic policy DSL, dynamic plugins, `.so`, WASM и runtime marketplace.
- Внешняя проверка паспортов/счетов и BIK lookup.
- PERSON/ADDRESS, secrets и aggressive date recognition.

## Планируемые задачи OpenSpec

- [ ] Уточнить supported formats и required context в specs/design.
- [ ] Реализовать PASSPORT detector и negative corpus.
- [ ] Реализовать BANK_ACCOUNT detector и documented validation limits.
- [ ] Реализовать contextual DATE_OF_BIRTH detector.
- [ ] Реализовать CustomRegexpDetector config/constructor.
- [ ] Интегрировать custom findings с resolver/masking.
- [ ] Добавить boundary, zero-width, concurrency и unsafe-error tests.
- [ ] Провести public API review и зафиксировать принятые изменения ADR.

## Acceptance

1. Обычные даты и неконтекстные numeric strings не становятся PII.
2. Supported passport/account forms перечислены точно; неподдерживаемые формы не
   выдаются за надёжно валидированные.
3. Пользователь может добавить custom string entity regexp и custom Go Detector,
   после чего оба проходят общий end-to-end pipeline.
4. Невалидный regexp/config/finding возвращает безопасную ошибку без input value.
5. После API review дальнейшие M4–M7 не требуют breaking core change без нового ADR.

## Verification

```bash
go test ./... -run 'Test(Passport|BankAccount|DateOfBirth|Regex|Custom)'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Каждый regexp/structured fuzz target запускается отдельно по точному package/name,
которые должны быть записаны в OpenSpec tasks и Composer packet.

## Exit evidence

- Archived deltas `structured-pii` и `custom-detection`.
- Public API review note/ADR и working custom entity example.
- Полная per-entity structured corpus summary.
