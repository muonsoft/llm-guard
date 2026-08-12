## Why

После завершения structured PII scope библиотеке нужен первый rule-based linguistic detector: консервативный RU PERSON с точными UTF-8 byte spans и контролируемым false-positive profile. M4 переносит PERSON-часть MVP-плана в канонический контракт, используя уже принятые в R0 tokenizer, bounded matcher и reference baseline.

## What Changes

- Добавляется pure-Go `PersonDetector` для согласованных двух- и трёхкомпонентных ФИО, распространённых склонённых форм и вариантов с инициалами.
- Добавляется внутренний минимальный NLP runtime поверх `github.com/muonsoft/go-razdel`, сохраняющий исходные byte offsets и не экспортирующий morphology/grammar API.
- Добавляются versioned positive/negative/differential corpus evaluation, concurrency regression coverage и воспроизводимый quality report.
- PERSON интегрируется с существующими `Guard`, resolver, masking и restore без изменения их публичных контрактов.
- Уточняется ограничение восстановления: исходная словоформа возвращается byte-for-byte, но библиотека не согласует её с изменённым LLM-контекстом.

## Capabilities

### New Capabilities

- `russian-person`: консервативное обнаружение RU PERSON, точные spans, bounded morphology/rules, quality boundary и end-to-end pipeline.

### Modified Capabilities

Нет: существующий `nlp-reference` уже задаёт pinned baseline, runtime boundary и intentional-difference policy, которые M4 применяет без изменения требований.

## Impact

- Production code: новый публичный detector и внутренние tokenizer/annotation/rule packages.
- Dependencies: добавляется разрешённый R0 pure-Go tokenizer `github.com/muonsoft/go-razdel` на зафиксированной revision; Python/reference dependencies не входят в Go module graph.
- Tests/data/docs: расширяется PERSON corpus evidence и документируется limitation Mask/Restore; реальные PII и неаудированные словари не добавляются.
- Breaking changes отсутствуют.
