## 1. Context и shared detector primitives

- [x] 1.1 Реализовать bounded same-line RU context matching с case-insensitive markers и безопасными UTF-8 byte boundaries.
- [x] 1.2 Добавить общие helpers для точной numeric normalization, homogeneous digits и malformed numeric extensions без включения values в errors.
- [x] 1.3 Зафиксировать existing resolver fallback: все built-in entities имеют приоритет выше unknown custom string entity; добавить overlap regressions.

## 2. PASSPORT detector

- [x] 2.1 Реализовать immutable `NewPassportDetector` с формами `NNNN NNNNNN` и `NN NN NNNNNN`, обязательным passport context и точным numeric span.
- [x] 2.2 Отклонять separated series/number, marker-free, malformed, mixed-separator, multiline-context и embedded numeric passport candidates.
- [x] 2.3 Добавить PASSPORT positive, negative, malformed, punctuation, Unicode boundary и concurrent tests.

## 3. BANK_ACCOUNT detector

- [x] 3.1 Реализовать immutable `NewBankAccountDetector` для compact 20 digits и пяти групп по четыре цифры с обязательным bank-account context.
- [x] 3.2 Реализовать optional same-line BIK extraction и checksum validation; при отсутствии БИК сохранить документированный context-first fallback.
- [x] 3.3 Отклонять unrelated contract/reference numbers, invalid length, repeated digits, malformed/mixed separators, numeric extensions и failed available checksum.
- [x] 3.4 Добавить BANK_ACCOUNT positive, BIK-valid/invalid, fallback, negative, malformed, Unicode boundary и concurrent tests.

## 4. DATE_OF_BIRTH detector

- [x] 4.1 Реализовать immutable `NewDateOfBirthDetector` для numeric `DD.MM.YYYY`/`DD/MM/YYYY` с calendar validation и обязательным RU birth-context.
- [x] 4.2 Добавить textual `D <русский месяц в родительном падеже> YYYY [года]` с точным date-only span.
- [x] 4.3 Отклонять ordinary dates, contract dates, impossible/malformed dates, ambiguous years, multiline context и embedded date-like tokens.
- [x] 4.4 Добавить DATE_OF_BIRTH positive, negative, malformed, punctuation, Unicode boundary и concurrent tests.

## 5. Custom regexp и custom Go integration

- [x] 5.1 Реализовать public `RegexDetectorConfig`, immutable `CustomRegexpDetector` и `NewCustomRegexpDetector` с compile-once RE2 pattern.
- [x] 5.2 Валидировать name, entity, pattern и finite confidence; возвращать `ErrInvalidConfig` с фиксированным безопасным reason без caller values или wrapped regexp error.
- [x] 5.3 Реализовать full-match non-overlapping findings, exact UTF-8 byte spans, context cancellation и безопасный skip zero-width matches без implicit boundaries.
- [x] 5.4 Добавить regexp constructor/matching/zero-width/boundary/concurrency tests и safe-error assertions для чувствительных config values.
- [x] 5.5 Добавить end-to-end tests для custom regexp и внешнего Go `Detector` через `Detect → Resolve → Mask → Restore`, включая built-in overlap priority и invalid finding safety.

## 6. Corpus, fuzz и public API evidence

- [x] 6.1 Расширить mixed structured Guard/round-trip test тремя новыми built-in detectors.
- [x] 6.2 Расширить per-entity corpus evaluation PASSPORT, BANK_ACCOUNT и DATE_OF_BIRTH cases с отдельными positive/negative/malformed и false-positive/false-negative counts.
- [x] 6.3 Расширить package `.` target `FuzzStructuredDetectorsInvariants` новыми built-ins и seeds.
- [x] 6.4 Добавить package `.` target `FuzzCustomRegexpDetectorInvariants` для UTF-8 spans, zero-width safety, deterministic Detect/Resolve и отсутствия panic.
- [x] 6.5 Добавить public example custom string entity и обновить README supported structured forms/validation limits без обещания внешней валидации.
- [x] 6.6 Провести public API review M0–M3 и зафиксировать accepted extension/breaking-change boundary в `docs/adr/0004-mvp-public-api-boundary.md`.

## 7. Verification

- [x] 7.1 Выполнить focused `go test ./... -run 'Test(Passport|BankAccount|DateOfBirth|Regex|Custom)'`.
- [x] 7.2 Выполнить отдельно `go test . -run '^$' -fuzz '^FuzzStructuredDetectorsInvariants$' -fuzztime=2s`.
- [x] 7.3 Выполнить отдельно `go test . -run '^$' -fuzz '^FuzzCustomRegexpDetectorInvariants$' -fuzztime=2s`.
- [x] 7.4 Выполнить broad `go test ./...`, `go vet ./...` и `go test -race ./...`.
- [x] 7.5 Сформировать verification report, проверить полноту requirements/scenarios и выполнить strict validation change перед OpenSpec verify/sync/archive.
