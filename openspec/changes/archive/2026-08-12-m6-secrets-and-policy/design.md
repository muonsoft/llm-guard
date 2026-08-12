## Context

См. `proposal.md` — Why. Existing core уже предоставляет concurrent `Guard.Detect`, deterministic `Resolve` и reversible `Guard.Mask`/`Restore`; secret entity constants присутствуют как зарезервированные, но detectors и action layer отсутствуют. M6 пересекает public API, detection, overlap resolution, masking/error semantics и security tests, оставаясь pure Go library без adapters.

Provider formats являются меняющимся внешним контрактом. Snapshot M6 фиксируется на 2026-08-12 по официальным GitHub/GitLab/AWS sources; OpenAI-like pattern явно считается conservative heuristic, потому что официальный exact shape не публикуется как стабильный контракт.

## Goals / Non-Goals

**Goals:**

- дать каждому secret family отдельный immutable detector с exact spans и safe metadata;
- ввести минимальную immutable action configuration с secure default block для secrets;
- сохранить единый resolver/masking pipeline, byte-offset contract и concurrent use одного Guard;
- сделать provider pattern maintenance auditable через version note и versioned synthetic corpus.

**Non-Goals:**

- cryptographic JWT signature/key validation, payload claims inspection или live provider verification;
- entropy/semantic generic-secret detection;
- YAML/rego DSL, dynamic rules, tenants, audit sink, rotation/revocation либо persistence;
- password query parameters и arbitrary/custom DSN schemes.

## Decisions

### 1. Один public detector на semantic family

Выбор: constructors `NewJWTDetector`, `NewPEMPrivateKeyDetector`, `NewAPIKeyDetector` и `NewDSNDetector`, каждый реализует существующий `Detector`. Family-specific files и tests остаются небольшими, а caller сам регистрирует нужные detectors.

Rationale: единый mega-detector скрыл бы precision/maintenance boundaries и усложнил бы отключение noisy family. Отдельные detectors проходят существующую validation/concurrency модель без нового registry.

### 2. JWT validation не читает claims

Выбор: scanner требует ровно три bounded base64url segments; локально декодируется и парсится только header object для непустого string `alg` и запрета `none`. Payload проверяется как непустая base64url shape, но claims не декодируются и не интерпретируются; signature segment также проверяется syntactically.

Rationale: проверка header резко снижает false positives, а отказ от claims parsing уменьшает sensitive-data surface. Signature verification невозможна без key material и остаётся вне scope.

### 3. PEM проверяется синтаксически без key parsing

Выбор: использовать standard-library PEM/base64 facilities для exact supported private labels, matching footer и non-empty body; не вызывать x509/OpenSSH parsers.

Rationale: encrypted и provider-specific private keys могут быть корректными без доступного password/algorithm support. Полный cryptographic parse дал бы false negatives и не нужен для masking.

### 4. Provider patterns — малый snapshot, а не generic entropy

Выбор: `NewAPIKeyDetector` распознаёт только version-note shapes и строгие token boundaries. GitHub snapshot включает documented `ghp_` и `github_pat_`; GitLab — `glpat-`; AWS — documented 20-character `AKIA`/`ASIA` access-key IDs; OpenAI-like — `sk-`/`sk-proj-` с conservative alphabet/minimum body, помеченный heuristic. Exact regex/lengths и synthetic examples живут рядом с source note и corpus.

Rationale: prefix-only recognition слишком шумное, а широкий набор быстро устаревает. Альтернатива с live metadata/provider API нарушила бы offline/pure-Go boundary.

### 5. DSN — allowlist schemes плюс standard URL parsing

Выбор: сначала выделять bounded URL-like candidates, затем `net/url` проверяет absolute supported scheme, host и `User.Password()` с непустыми username/password. HTTP(S), passwordless и query-only credentials отвергаются.

Rationale: одна regexp для DSN плохо обрабатывает percent encoding, IPv6 и query. Standard parser даёт deterministic conservative semantics без dependency; allowlist удерживает обычные URL вне secret family.

### 6. Policy хранится как immutable configuration Guard

Выбор: public `Action` и options `WithSecretAction(action)` / `WithEntityAction(entity, action)`. Construction валидирует action и duplicate exact overrides, копирует map; precedence: exact entity → secret-family override/default block → default mask. Все четыре secret entity классифицируются внутренней pure function.

Rationale: отдельный policy interface/DSL преждевременен. Options сохраняют established construction pattern и не добавляют mutable state. `allow` нужен для локального исключения; custom entities по умолчанию mask для backward compatibility.

### 7. Block оценивается до entropy и возвращает safe typed error

Выбор: Mask выполняет Detect → Resolve → classify. При любом block сразу возвращается zero `MaskResult` и `BlockError`, проверяемый через `ErrBlocked`; exported error data не содержит entity/span/count sequence. Только при отсутствии block формируются replacements для mask findings; allow findings остаются в `MaskResult.Findings`.

Rationale: partial masked output рядом с blocked credential легко ошибочно отправить дальше. Ранний block исключает token allocations/entropy reads и сужает leakage surface. Findings в успешном result остаются полным resolved evidence, как и до policy.

### 8. Secret priority выше URL/EMAIL

Выбор: `CONNECTION_STRING` получает наивысший secret priority, остальные validated secret entities — следующий уровень выше URL. Resolver остаётся общим и не знает detector implementation details.

Rationale: DSN часто синтаксически является URL и содержит email-like userinfo; сохранение внешнего URL потеряло бы secret classification и могло применить default mask вместо block. Полный existing tie-break остаётся без изменений.

## Risks / Trade-offs

- [Provider изменит token shape] → date-pinned source note, exact corpus и explicit update procedure; unknown versions остаются undetected вместо расширения heuristic.
- [Synthetic prefix-shaped text даст false positive] → strict boundaries, documented lengths/alphabet и secure default block; caller может exact-entity allow, понимая риск.
- [JWT payload shape окажется ложным совпадением] → header JSON/alg validation и fuzz/malformed corpus; claims намеренно не читаются.
- [URL candidate extraction обрежет punctuation либо percent-encoding] → table tests для punctuation/IPv6/query и `net/url` как второй validation stage.
- [Block error случайно раскроет secret через wrapping/formatting] → отдельный typed error без raw cause/value и exhaustive `%v`/`%+v`/`%#v`/JSON-adjacent tests.

## Migration Plan

Изменение additive для callers, которые не регистрируют secret detectors и не задают action options: существующие PII остаются default mask. Callers, добавляющие secret detector без override, получают новый secure default block; для прежнего mask-like поведения они явно передают `WithSecretAction(ActionMask)`. Rollback состоит в удалении новых detector/options usage; persistence и schema migration отсутствуют.
