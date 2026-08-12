## 1. Core и errors

- [x] 1.1 Добавить safe sentinels/typed errors для resolver, namespace source и invalid TokenSet без sensitive fields или error text.
- [x] 1.2 Расширить immutable Guard конфигурацией secure random source и serialized injectable `WithRandomSource` без изменения существующего Detect API.

## 2. EMAIL detector

- [x] 2.1 Реализовать public immutable `NewEmailDetector` через compiled candidate regexp и explicit local/domain validation.
- [x] 2.2 Реализовать Unicode-aware outer boundaries, context cancellation и canonical EMAIL finding metadata/byte spans.
- [x] 2.3 Добавить positive/negative corpus tests для common mailbox, punctuation, malformed forms, URL/placeholder-like text и mixed Unicode offsets.

## 3. Finding resolver

- [x] 3.1 Реализовать public pure `Resolve` с повторной UTF-8/finding validation без мутации input slice.
- [x] 3.2 Реализовать internal entity priority и полный deterministic selection key для deduplication, nesting и overlaps.
- [x] 3.3 Реализовать stable textual output order и tests для exact duplicates, adjacent spans, EMAIL/custom priority, equal-priority ties и permutations.

## 4. Reversible masking

- [x] 4.1 Добавить public `MaskResult` и opaque caller-owned `TokenSet` с private ordered mappings.
- [x] 4.2 Реализовать 128-bit namespace generation, фиксированный token envelope, occurrence counters, collision retry до 32 попыток и safe entropy errors.
- [x] 4.3 Реализовать `Guard.Mask` как Detect → Resolve → reverse byte-order replacement, включая no-findings и repeated-value semantics.
- [x] 4.4 Реализовать exact one-pass `Guard.Restore` для known tokens, repeated placeholders, unknown/mutated/cross-TokenSet tokens и nil TokenSet.
- [x] 4.5 Реализовать redacted `String`/`GoString` и safe `encoding/json` behavior для TokenSet и MaskResult.
- [x] 4.6 Добавить collision, random-source failure, mixed Unicode, repetition, recursive-looking value, context и concurrent Mask/Restore tests.

## 5. Property и fuzz verification

- [x] 5.1 Добавить `FuzzEmailDetectorBoundaries` с invariant валидных UTF-8 EMAIL spans и отсутствия panic.
- [x] 5.2 Добавить `FuzzResolveInvariants` с invariants valid sorted non-overlapping deterministic output.
- [x] 5.3 Добавить `FuzzMaskRestoreRoundTrip` с invariant `Restore(Mask(text)) == text` для неизменённого masked text.
- [x] 5.4 Выполнить отдельно `go test . -run '^$' -fuzz '^FuzzEmailDetectorBoundaries$' -fuzztime=20s`, `go test . -run '^$' -fuzz '^FuzzResolveInvariants$' -fuzztime=20s` и `go test . -run '^$' -fuzz '^FuzzMaskRestoreRoundTrip$' -fuzztime=20s`.

## 6. Docs и broad checks

- [x] 6.1 Добавить компилируемый embedded example `Detect → Mask → Restore` и обновить README status/usage без provider-specific API.
- [x] 6.2 Выполнить focused `go test ./... -run 'TestEmail|TestResolve|TestMask|TestRestore|TestToken'`.
- [x] 6.3 Выполнить broad `go test ./...`, `go vet ./...` и `go test -race ./...`.
