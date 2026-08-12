## Why

После консервативного RU PERSON detector библиотеке нужен второй linguistic detector из MVP: композиционный RU ADDRESS, который маскирует только согласованный полный адрес и не превращает одиночные географические названия в PII. M5 переносит ADDRESS scope из `docs/light_llm_guard_go_mvp_plan.md` в канонический контракт и закрывает вложенный конфликт ADDRESS/PERSON.

## What Changes

- Добавляется immutable pure-Go `AddressDetector` для распространённых русских settlement, street, house, building и apartment forms с точными UTF-8 byte spans.
- Вводится консервативная composition matrix: finding требует как минимум street+house; settlement и расширенные building parts усиливают либо расширяют уже принятую композицию, но не образуют ADDRESS самостоятельно.
- Добавляются versioned positive, negative и ambiguous corpus, offline differential comparison с pinned Natasha fixtures и отдельный quality report.
- ADDRESS интегрируется с общим Detect/Resolve/Mask/Restore pipeline, включая concurrency и punctuation/Unicode regression coverage.
- Resolver policy уточняется: полный ADDRESS имеет приоритет над вложенным PERSON, в том числе для `ул. Академика Сахарова, 10`.

## Capabilities

### New Capabilities

- `russian-address`: консервативное обнаружение композиционных RU ADDRESS, точные spans, supported parts, quality boundary и end-to-end pipeline.

### Modified Capabilities

- `finding-resolution`: явный deterministic priority contract для ADDRESS над вложенным PERSON.

## Impact

- Production code: новый публичный detector и внутренние immutable address annotations/composition rules поверх существующего tokenizer runtime.
- Public API: добавляется constructor detector без изменения существующих interfaces или default Guard configuration.
- Dependencies: новые production dependencies и runtime data не требуются; Natasha остаётся development-only reference baseline.
- Tests/data/docs: добавляются address corpus evaluation, resolver overlap regression, Mask/Restore example и quality report. Breaking changes отсутствуют.
