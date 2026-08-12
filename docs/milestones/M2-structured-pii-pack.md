# M2 — Основной structured PII pack

| Поле | Значение |
|---|---|
| Change | `m2-structured-pii-pack` |
| Capability | `structured-pii` |
| Зависимости | M1 |
| Default variant | C |
| Результат | Практически полезное покрытие основных детерминированных PII |

## Outcome

Расширить готовый end-to-end pipeline детекторами RU PHONE, IP, URL, INN, SNILS и
BANK_CARD. Каждый тип проходит candidate → normalize → validate → finding и
маскируется/восстанавливается существующим core без специального кода в Guard.

## Scope

- PHONE: RU `+7`/`8`, скобки/разделители и conservative international baseline.
- IPv4 и IPv6 с validation стандартной библиотекой.
- URL, включая query/credentials boundaries и overlap с EMAIL.
- INN физлиц/юрлиц с checksum.
- SNILS с normalization и checksum.
- BANK_CARD с separators, Luhn и false-positive boundaries.
- Общие reusable candidate/normalization helpers только там, где они сокращают
  дублирование без создания framework.
- Per-entity positive/negative/malformed/Unicode corpus и mixed overlap cases.
- Resolver priorities для EMAIL/URL и structured numeric entities.

## Out of scope

- PASSPORT, BANK_ACCOUNT, DATE_OF_BIRTH, PERSON, ADDRESS и secrets.
- High-entropy detection, external validation и network calls.
- Агрессивная классификация любой длинной цифровой строки.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать formats, validators и conservative exclusions по entity.
- [ ] Реализовать PHONE detector и corpus.
- [ ] Реализовать IPv4/IPv6 detector и corpus.
- [ ] Реализовать URL detector и EMAIL overlap cases.
- [ ] Реализовать INN и SNILS checksum detectors.
- [ ] Реализовать BANK_CARD Luhn detector.
- [ ] Обновить resolver priorities и mixed corpus.
- [ ] Добавить malformed/property/fuzz и end-to-end tests по каждому entity.

## Acceptance

1. Все шесть entity types имеют positive, negative и malformed corpus coverage.
2. INN, SNILS и BANK_CARD не создают finding при неверной checksum.
3. PHONE не маскирует произвольные длинные числа, IP использует semantic parsing,
   URL и EMAIL разрешают overlaps детерминированно.
4. Каждый accepted finding проходит существующий Mask/Restore round-trip.
5. Per-entity evaluation выводит отдельные counts/quality, не только aggregate.

## Verification

```bash
go test ./... -run 'Test(Phone|IP|URL|INN|SNILS|BankCard|Structured|Mixed)'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Для каждого добавленного fuzz target FF/tasks фиксирует точный package и запускает
его отдельно: `go test <package> -run '^$' -fuzz '^<exact-target-name>$'
-fuzztime=20s`.

## Exit evidence

- Archived structured-pii delta, merged без потери M1 EMAIL scenarios.
- Corpus/evaluation evidence по каждому entity.
- Review подтверждает отсутствие raw identifiers в errors и diagnostics.
