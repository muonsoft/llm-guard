## Context

M1 уже предоставляет immutable `Detector`, concurrent-safe `Guard`, validated UTF-8 byte spans, deterministic `Resolve` и generic reversible masking. См. `proposal.md` и delta specs. M2 добавляет несколько разных lexical grammars, но должен сохранить единый public extension mechanism, отсутствие matched values в findings/errors и offline pure-Go execution.

## Goals / Non-Goals

**Goals:**

- Реализовать candidate → normalize → validate → finding для всего M2 pack с небольшими shared helpers только для одинаковых boundary/normalization операций.
- Зафиксировать conservative форматы и semantic/checksum validation до написания regexp.
- Сохранить detectors immutable и пригодными для parallel use одного Guard.
- Доказать byte-span, overlap, false-positive и Mask/Restore invariants общим corpus и fuzz target.

**Non-Goals:**

- PASSPORT, BANK_ACCOUNT, DATE_OF_BIRTH, PERSON, ADDRESS, secrets и custom regexp detector.
- Полный RFC URL/telephone parser, IDN/IRI, IPv6 zone identifiers, `tel:` URI и external validation.
- Public resolver policy или configuration DSL.
- Observer/metrics production API; per-entity quality в M2 является verification evidence.

## Decisions

### 1. Каждый entity остаётся отдельным immutable detector

Публичные constructors `NewPhoneDetector`, `NewIPDetector`, `NewURLDetector`, `NewINNDetector`, `NewSNILSDetector` и `NewBankCardDetector` возвращают существующий `Detector` и регистрируются через `WithDetector`. Это сохраняет явный caller choice и не закрепляет premature `WithDefaultPIIDetectors` до M3. Один monolithic detector отвергнут: у entities разные validators, confidence и corpus boundaries, а отдельные detectors проще комбинировать и оценивать.

### 2. Candidate regexp ограничивает blast radius, validator принимает окончательное решение

Regexp ищет только bounded ASCII candidates, после чего detector проверяет outer rune boundaries, нормализует разрешённые separators и вызывает entity validator. Findings покрывают исходную lexical форму, а normalized value живёт только в локальной переменной и не попадает в `Finding`, error или diagnostic. Shared helpers ограничиваются digit collection, rune-aware boundaries и safe finding construction; общий detector framework не вводится.

Альтернатива «один permissive regexp на entity» отвергнута из-за checksum/semantic false positives. Полный streaming lexer также не нужен для небольшого MVP grammar.

### 3. PHONE поддерживает явные country/trunk prefixes

RU candidate принимается только как `+7` или trunk `8` плюс десять national digits; разрешаются распространённые пробелы/дефисы и одна согласованная пара скобок вокруг трёхзначного area code. International baseline требует leading `+`, 8–15 total digits, по крайней мере один separator либо non-RU country code и корректные outer boundaries. Unprefixed local numbers и произвольные contiguous long numbers отвергаются.

Такой профиль предпочитает false negatives агрессивному распознаванию order/account identifiers как PHONE. Контекстные слова и country metadata не добавляются, потому что это потребовало бы policy/NLP вне M2.

### 4. IP и URL подтверждаются standard-library parsers

IP candidate перед finding проверяется через `net.ParseIP`; IPv4 дополнительно требует ровно четыре canonical decimal octets без ambiguous leading zeros, IPv6 — наличие colon и отсутствие zone suffix. Brackets используются только как boundary для endpoint-form IPv6 и не входят в span.

URL detector ограничен absolute `http`/`https`, парсит candidate через `net/url`, требует non-empty hostname и либо dotted DNS-like host, либо semantically valid IP. Он сохраняет userinfo, port, path, query и fragment, но trim’ит только однозначную trailing sentence punctuation с balanced-delimiter проверкой. Relative URL, unsupported scheme и whitespace не принимаются. Собственный полный URL grammar отвергнут как более рискованный и несовместимый со стандартным Go behavior.

### 5. Numeric identifiers используют нормативные checksums и строгую форму

INN принимает только 10 или 12 contiguous digits, отвергает homogeneous repeated digits и проверяет соответствующие weighted checksum digits. SNILS принимает compact 11 digits или ровно `XXX-XXX-XXX YY`, отвергает номера ниже `001001998`, для которых checksum исторически не определяет валидность, и применяет официальный modulo-101 rule. BANK_CARD принимает 13–19 digits, optional consistent spaces или hyphens, отвергает homogeneous digits и применяет Luhn.

Detectors не требуют textual labels: checksum и strict boundaries являются контрактом M2. Более сильная context policy может быть добавлена без изменения pipeline, но ослабление checksum не допускается.

### 6. URL является enclosing overlap, structured checksums побеждают generic matches

Resolver поднимает URL выше EMAIL, когда spans пересекаются, чтобы full URL с credentials/query не был частично замаскирован и не оставлял чувствительный remainder. EMAIL остаётся выше unknown custom entities. BANK_CARD, SNILS и INN также выше default custom priority; остальные tie-break rules и stable output order не меняются.

Альтернатива EMAIL > URL отвергнута для enclosing URL: она заменяла бы только mailbox fragment и оставляла остальную URL PII. Detector order не используется как policy, поэтому результат остаётся permutation-independent.

### 7. Corpus и один fuzz target дают раздельное evidence

Table-driven corpus содержит positive, negative и malformed cases по каждой entity плюс Unicode и mixed overlaps. Evaluation test вычисляет expected/detected/false-positive/false-negative counts отдельно по entity и делает mismatch test failure, не вводя runtime metrics API. Один package-level `FuzzStructuredDetectorsInvariants` проверяет отсутствие panic, валидные UTF-8 spans, отсутствие raw values в formatting и deterministic Detect/Resolve на произвольном input; отдельный existing Mask/Restore fuzz продолжает доказывать generic round trip.

Один fuzz target выбран вместо шести почти одинаковых harnesses, потому что invariant и package boundary общие, а seed corpus остаётся per-entity.

## Risks / Trade-offs

- [Regexp может вырезать валидную часть из malformed длинного candidate] → проверять rune-aware outer boundaries до validator и держать malformed corpus с prefix/suffix extensions.
- [URL trailing punctuation неоднозначна] → trim’ить только фиксированный terminal set и учитывать balanced parentheses/brackets; спорные формы отвергать.
- [PHONE international baseline даёт false positives] → требовать явный `+`, E.164 digit bounds и conservative grouping/boundaries.
- [Luhn-valid случайное длинное число всё ещё возможно] → строгая длина, boundaries, consistent separators и homogeneous-digit exclusion; context scoring отложен.
- [URL > EMAIL меняет M1 overlap policy] → зафиксировать полную MODIFIED requirement и permutation/masking regression tests.

## Migration Plan

Изменение additive для public constructors и entity coverage. Единственное observable изменение существующего API — resolver теперь выбирает enclosing URL вместо overlapping EMAIL при наличии обоих findings; это синхронизируется в `finding-resolution` spec и покрывается regression test. Rollback удаляет новые detector files/constructors и возвращает прежнюю internal priority, не затрагивая persisted state или TokenSet format.
