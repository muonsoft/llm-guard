## 1. Core

- [x] 1.1 Добавить public secret `EntityType` constants и стабильные detector names, сохранив строковую расширяемость entity model.
- [x] 1.2 Добавить строковый `Action` с `allow`, `mask`, `block` и unit tests публичного контракта.
- [x] 1.3 Реализовать immutable `WithSecretAction` и `WithEntityAction` options с валидацией nil/empty/unknown/duplicate configuration.
- [x] 1.4 Реализовать deterministic action resolution: exact entity, затем secret-family override/default block, затем default mask для PII/custom.
- [x] 1.5 Поднять priorities всех secret entities над URL/EMAIL/custom, сохранив существующие resolver tie-breaks.
- [x] 1.6 Добавить resolver tests для DSN↔URL/EMAIL, secret↔custom, candidate permutations и stable output.

## 2. Detectors

- [x] 2.1 Реализовать bounded JWT candidate scan и header-only structural validation без claims parsing/logging.
- [x] 2.2 Добавить JWT unit cases для valid, `alg:none`, malformed base64/JSON, segment count, boundaries и exact byte spans.
- [x] 2.3 Реализовать PEM private-key detector для пяти разрешённых labels с matching footer и syntactically valid non-empty body.
- [x] 2.4 Добавить PEM tests для supported labels, public/certificate negatives, mismatched/missing footer, malformed body и surrounding text.
- [x] 2.5 Реализовать version-pinned provider token detector для GitHub, GitLab, OpenAI-like и AWS shapes со строгими boundaries.
- [x] 2.6 Добавить provider positive/negative/malformed tests для каждого supported shape, truncation, embedding и invalid alphabets.
- [x] 2.7 Реализовать DSN candidate scan и conservative `net/url` validation для allowlisted schemes с non-empty username/password.
- [x] 2.8 Добавить DSN tests для percent-encoding, IPv6/query, punctuation, HTTP(S), passwordless, relative и query-only credential negatives.
- [x] 2.9 Добавить context-cancellation и concurrent deterministic tests для каждого secret detector.

## 3. Masking

- [x] 3.1 Встроить action classification после Resolve и до namespace generation в `Guard.Mask`.
- [x] 3.2 Реализовать allow path с полным resolved `Findings`, неизменёнными bytes и mappings только для mask findings.
- [x] 3.3 Реализовать ранний block path с zero `MaskResult` и без entropy read/partial replacement.
- [x] 3.4 Сохранить collision-safe namespace, reverse-byte replacement и Restore semantics для subset mask findings.
- [x] 3.5 Добавить mixed allow/mask/block, all-allow, default-secret-block, explicit-secret-mask и existing PII regression tests.

## 4. Audit и errors

- [x] 4.1 Добавить `ErrBlocked` и typed safe block error, поддерживающий `errors.Is`/`errors.As` без raw cause или occurrence metadata.
- [x] 4.2 Проверить `%v`, `%+v`, `%#v`, JSON-adjacent и wrapped error surfaces на отсутствие raw secrets, JWT payload, PEM body, DSN password и tokens.
- [x] 4.3 Проверить, что detector validation/errors не сохраняют raw candidate или decoded values.

## 5. Tests

- [x] 5.1 Добавить versioned synthetic positive/negative/malformed corpus с family/shape metadata и exact byte spans без live credentials.
- [x] 5.2 Добавить общий corpus test через Detect→Resolve и per-family zero mandatory false-positive/false-negative assertions.
- [x] 5.3 Добавить fuzz targets `FuzzJWTDetector`, `FuzzPEMDetector`, `FuzzAPIKeyDetector`, `FuzzDSNDetector` с invariants для panic, spans и safe errors.
- [x] 5.4 Запустить focused suite `go test ./... -run 'Test(JWT|PEM|APIKey|Secret|DSN|Policy|Block)'`.
- [x] 5.5 Запустить каждый fuzz target отдельно: `go test . -run '^$' -fuzz '^FuzzJWTDetector$' -fuzztime=2s`, `go test . -run '^$' -fuzz '^FuzzPEMDetector$' -fuzztime=2s`, `go test . -run '^$' -fuzz '^FuzzAPIKeyDetector$' -fuzztime=2s`, `go test . -run '^$' -fuzz '^FuzzDSNDetector$' -fuzztime=2s`.
- [x] 5.6 Запустить broad checks `go test ./...`, `go vet ./...`, `go test -race ./...`.

## 6. Docs

- [x] 6.1 Добавить `docs/secret-patterns.md` со snapshot date, official source URLs, exact supported shapes, OpenAI-like heuristic caveat и update procedure.
- [x] 6.2 Обновить package examples/README для явного default block и `WithSecretAction(ActionMask)` без реальных credentials.
- [x] 6.3 Проверить `go.mod`/imports: нет provider SDK, network, logging, persistence или policy-framework dependencies.
- [x] 6.4 Выполнить `openspec validate m6-secrets-and-policy --strict --no-interactive` и зафиксировать verification evidence для каждой requirement/scenario.
