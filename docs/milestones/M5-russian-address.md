# M5 — Композиционный RU ADDRESS

| Поле | Значение |
|---|---|
| Change | `m5-russian-address` |
| Capabilities | `russian-address`, `finding-resolution` |
| Зависимости | M4 |
| Default variant | C |
| Результат | Композиционные адреса маскируются целиком без nested PERSON leakage |

## Outcome

Реализовать минимальный pure-Go address extractor по R0 strategy и product-level
compositional policy. Самостоятельный город или регион не считается ADDRESS;
street+house и более полные композиции считаются.

## Scope

- Приоритетные parts: settlement, street/prospect/lane/highway, house, corpus,
  building/structure, apartment; postal index только если предусмотрен design.
- RU abbreviations, punctuation, reordered common address forms и byte spans.
- Compositional acceptance rules: минимум street+house либо более сильная
  согласованная комбинация.
- Отдельные extractor parts могут быть internal; публичный finding охватывает
  принятый полный адрес.
- Resolver priority `ADDRESS > PERSON`, когда PERSON вложен в название улицы.
- Positive, negative и ambiguous corpus, включая `ул. Академика Сахарова`.
- Differential comparison и documented intentional differences.
- End-to-end masking, concurrency и quality evaluation.

## Out of scope

- Одиночные Москва/регион/улица, geocoding и нормализация до ФИАС.
- Полное покрытие Natasha address grammar и generic location NER.
- Network calls, external address databases и organization extraction.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать supported address parts и composition matrix.
- [ ] Реализовать street/house и common abbreviation rules.
- [ ] Добавить settlement и extended building/apartment parts.
- [ ] Реализовать full-span composition и confidence/acceptance logic.
- [ ] Обновить resolver priority ADDRESS/PERSON.
- [ ] Добавить positive/negative/ambiguous corpus.
- [ ] Интегрировать AddressDetector с common pipeline.
- [ ] Выполнить differential, concurrency и quality report.

## Acceptance

1. Mandatory positive examples из concept дают один полный ADDRESS finding.
2. Самостоятельные города, регионы и названия улиц из negative corpus не
   маскируются как ADDRESS.
3. `ул. Академика Сахарова, 10` не оставляет отдельный PERSON finding внутри
   принятого ADDRESS.
4. Byte spans, deterministic resolution и Mask/Restore корректны на punctuation
   и Unicode variants.
5. Runtime остаётся pure Go и использует только лицензированные R0 dependencies.

## Verification

```bash
go test ./... -run 'Test(Address|Addr|Resolve.*Address|Mixed)'
go test -race ./... -run 'TestAddress'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Дополнительно запускаются R0 differential harness и address corpus evaluation.

## Exit evidence

- Archived `russian-address` и updated `finding-resolution` specs.
- Negative corpus подтверждает compositional policy.
- Review отдельно фиксирует ADDRESS/PERSON overlap result.
