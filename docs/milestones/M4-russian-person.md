# M4 — Консервативный RU PERSON

| Поле | Значение |
|---|---|
| Change | `m4-russian-person` |
| Capabilities | `russian-person`, при необходимости `nlp-reference` |
| Зависимости | M3 и R0 |
| Default variant | C |
| Результат | Rule-based PERSON detection для согласованного консервативного corpus |

## Outcome

Реализовать pure-Go PersonDetector по архитектуре, принятой в R0. Главная
product boundary — корректный PERSON span и контролируемый false-positive profile,
а не Natasha API compatibility или разбор факта на поля.

## Scope

- Tokenizer/morphology/rule runtime строго в объёме ADR R0.
- ФИО в порядках first-last, last-first, с отчеством и распространёнными
  склонёнными формами.
- Формы с инициалами `Петров И. С.` и `И. С. Петров`.
- Консервативная капитализация, punctuation и token boundary policy.
- Byte spans исходного UTF-8 текста без реконструкции offsets из нормализованного.
- Negative corpus: одиночные имена/фамилии, нарицательные совпадения, названия
  продуктов/проектов и street-like contexts.
- Differential comparison с pinned reference baseline, intentional differences.
- End-to-end resolver/mask/restore integration и concurrency tests.

## Out of scope

- Одиночные `Иван`/`Петров`, nicknames, organizations и generic NER.
- Coreference, cross-message identity, normalization и morphology-aware restore.
- Расширение NLP runtime за пределы фактически нужного PersonDetector.

## Планируемые задачи OpenSpec

- [ ] Перенести R0 decisions и PERSON scenarios в delta spec/design.
- [ ] Реализовать/adapt tokenizer с сохранением byte offsets.
- [ ] Реализовать минимальные morphology/dictionary contracts из ADR.
- [ ] Реализовать name rules, initials и supported word orders.
- [ ] Добавить declined-form positive corpus.
- [ ] Добавить conservative negative corpus и regression cases.
- [ ] Интегрировать PersonDetector с Guard/resolver/masking.
- [ ] Добавить differential, concurrency и quality report.

## Acceptance

1. Все mandatory golden forms из MVP concept детектируются с точными byte spans.
2. Согласованный negative corpus не создаёт aggressive single-name findings.
3. Intentional Natasha differences перечислены и обоснованы product safety, а
   reference revision воспроизводим.
4. PersonDetector immutable/concurrent-safe и не тянет Python/ML/external API.
5. PERSON проходит общий Mask/Restore; limitation морфологического согласования
   после restore документирован.

## Verification

```bash
go test ./... -run 'Test(Person|Name|Morph|Token)'
go test -race ./... -run 'TestPerson'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Также запускаются pinned differential harness и per-entity corpus evaluation,
определённые R0. Их точные команды переносятся в OpenSpec tasks и task packet.

## Exit evidence

- Archived `russian-person` spec и, если менялся harness, `nlp-reference` delta.
- Golden/negative/differential отчёт с точными reference versions.
- License inventory остаётся полным после добавления dictionaries/data.
