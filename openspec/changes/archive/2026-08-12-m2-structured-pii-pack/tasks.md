## 1. Core и shared detector primitives

- [x] 1.1 Зафиксировать public constructors и стабильные detector names для PHONE, IP_ADDRESS, URL, INN, SNILS и BANK_CARD без изменения существующего `Detector` contract.
- [x] 1.2 Реализовать минимальные rune-aware outer-boundary и ASCII digit-normalization helpers без хранения input fragments в errors или exported state.
- [x] 1.3 Обновить internal resolver priority так, чтобы enclosing URL побеждал overlapping EMAIL, а validated INN/SNILS/BANK_CARD — unknown custom numeric findings.
- [x] 1.4 Добавить permutation и nested-overlap regression tests для URL/EMAIL и structured numeric/custom cases.

## 2. PHONE detector

- [x] 2.1 Реализовать RU `+7`/`8` candidate grammar с разрешёнными скобками/разделителями и точными исходными spans.
- [x] 2.2 Реализовать conservative international `+` validation и exclusions для local-only, malformed, embedded и arbitrary long numeric candidates.
- [x] 2.3 Добавить PHONE positive, negative, malformed, punctuation и Unicode byte-offset corpus tests.

## 3. IP_ADDRESS и URL detectors

- [x] 3.1 Реализовать IPv4 candidate extraction с canonical-octet checks и semantic standard-library validation.
- [x] 3.2 Реализовать bare/bracketed IPv6 extraction с semantic validation и exclusions для malformed/zone/embedded candidates.
- [x] 3.3 Добавить IPv4/IPv6 positive, negative, malformed, punctuation и Unicode byte-offset corpus tests.
- [x] 3.4 Реализовать absolute HTTP(S) URL extraction и standard-library validation для DNS/IP host, port, userinfo, path, query и fragment.
- [x] 3.5 Реализовать conservative trailing punctuation/balanced delimiter handling и exclusions для relative, unsupported-scheme, whitespace и malformed URL.
- [x] 3.6 Добавить URL corpus и mixed URL/EMAIL overlap tests, включая credentials, path/query fragments и permutation-independent resolution.

## 4. INN, SNILS и BANK_CARD detectors

- [x] 4.1 Реализовать 10/12-digit INN boundary checks, homogeneous-digit exclusion и обе официальные checksum schemes.
- [x] 4.2 Добавить INN legal/individual positive cases, failed checksum, malformed length/separator и embedded-number tests.
- [x] 4.3 Реализовать compact/formatted SNILS normalization, legacy-range exclusion и modulo-101 checksum.
- [x] 4.4 Добавить SNILS positive, failed-checksum, legacy, malformed/mixed-separator и embedded-number tests.
- [x] 4.5 Реализовать 13–19 digit BANK_CARD normalization с consistent separators, homogeneous-digit exclusion и Luhn validation.
- [x] 4.6 Добавить BANK_CARD compact/formatted positive cases, failed Luhn, repeated digits, malformed separators и embedded-number tests.

## 5. Masking, audit и evaluation

- [x] 5.1 Добавить mixed Unicode end-to-end test, где все шесть entity проходят `Detect → Resolve → Mask → Restore` без entity-specific Guard logic.
- [x] 5.2 Проверить tests, что findings, standard formatting, JSON и returned errors/diagnostics не содержат raw PHONE, IP, URL, INN, SNILS или BANK_CARD values.
- [x] 5.3 Добавить per-entity corpus evaluation с отдельными expected, detected, false-positive и false-negative counts для positive/negative/malformed cases.
- [x] 5.4 Добавить concurrent all-detectors test для immutable Guard и deterministic result без data race.

## 6. Property, fuzz, docs и verification

- [x] 6.1 Добавить package `.` fuzz target `FuzzStructuredDetectorsInvariants` с per-entity seeds и invariants UTF-8 spans, deterministic Detect/Resolve и отсутствия panic.
- [x] 6.2 Выполнить отдельно `go test . -run '^$' -fuzz '^FuzzStructuredDetectorsInvariants$' -fuzztime=20s`.
- [x] 6.3 Обновить README/public example для явной регистрации structured detector pack без provider-specific API.
- [x] 6.4 Выполнить focused `go test ./... -run 'Test(Phone|IP|URL|INN|SNILS|BankCard|Structured|Mixed)'`.
- [x] 6.5 Выполнить broad `go test ./...`, `go vet ./...` и `go test -race ./...`.
- [x] 6.6 Выполнить `openspec validate --specs --strict --no-interactive` после синхронизации delta specs.
