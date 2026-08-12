# M1 — Reversible EMAIL vertical slice

| Поле | Значение |
|---|---|
| Change | `m1-reversible-email-slice` |
| Capabilities | `structured-pii`, `finding-resolution`, `reversible-masking` |
| Зависимости | M0 |
| Default variant | C |
| Результат | EMAIL проходит полный `Detect → Resolve → Mask → Restore` |

## Outcome

Получить первый end-to-end embedded guard: EMAIL детектируется с корректным byte
span, конфликты разрешаются детерминированно, sensitive fragment заменяется
collision-safe placeholder и точно восстанавливается caller-owned `TokenSet`.

## Scope

- Консервативный EMAIL detector и positive/negative corpus.
- Общий deterministic resolver: validation, deduplication, overlap/nesting,
  configurable internal priority и stable output order.
- `Mask`, `MaskResult`, opaque `TokenSet`, placeholder namespace и counters.
- Namespace через cryptographically secure randomness с injectable source для tests.
- Reverse-order replacement по byte offsets без повреждения Unicode.
- Exact restore только tokens данного `TokenSet`; unknown/mutated tokens не
  раскрывают data и обрабатываются согласно spec.
- Повтор одного sensitive value и повтор placeholder имеют определённую семантику.
- Safe `String`/`GoString`; mappings не экспортируются и автоматически не
  сериализуются.
- Unit, property/fuzz, collision и mixed Unicode tests.

## Out of scope

- Другие PII types, audit observer, metrics и policy actions.
- Persistent/session storage, cross-TokenSet restore и morphology correction.
- Попытки восстанавливать изменённый моделью placeholder эвристически.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать resolver, token identity и restore semantics.
- [ ] Реализовать EMAIL candidate matching и boundary rules.
- [ ] Реализовать resolver и priority model.
- [ ] Реализовать opaque TokenSet и collision-safe placeholder generation.
- [ ] Реализовать Mask с Unicode-safe reverse replacements.
- [ ] Реализовать exact Restore и miss behavior.
- [ ] Добавить corpus, overlap, collision и repetition tests.
- [ ] Добавить fuzz invariants resolver и round-trip.
- [ ] Добавить embedded usage example.

## Acceptance

1. `Restore(Mask(text)) == text` для неизменённого masked text, включая Unicode,
   повторы и текст с placeholder-like fragments.
2. Resolved findings валидны, отсортированы и не пересекаются; одинаковый input
   всегда даёт одинаковое разрешение конфликтов.
3. Restore раскрывает только точные tokens переданного `TokenSet`, неизвестные
   tokens не подменяются.
4. Formatting, errors и стандартный JSON path не раскрывают TokenSet values.
5. Public core API остаётся provider-agnostic и concurrent-safe.

## Verification

```bash
go test ./... -run 'TestEmail|TestResolve|TestMask|TestRestore|TestToken'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Каждый конкретный fuzz target запускается отдельно в его package:

```bash
go test <package> -run '^$' -fuzz '^<exact-target-name>$' -fuzztime=20s
```

FF/tasks и task packet обязаны заменить placeholders точными значениями; нельзя
считать milestone зелёным, если команда не запустила target.

## Exit evidence

- Archived change и три canonical capability specs.
- Компилируемый end-to-end example.
- Fuzz/race evidence и review sensitive-data surfaces.
