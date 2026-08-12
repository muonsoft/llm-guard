# M6 — Basic secrets и minimal action policy

| Поле | Значение |
|---|---|
| Change | `m6-secrets-and-policy` |
| Capabilities | `secret-detection`, `minimal-policy` |
| Зависимости | M3 |
| Default variant | C |
| Результат | Основные secrets безопасно mask/block через общий Guard pipeline |

## Outcome

Добавить консервативные deterministic secret detectors и минимальный action layer,
не превращая core в policy engine. PII по умолчанию маскируется; secrets получают
явно настроенное `mask` или `block` поведение.

## Scope

- JWT с structural validation без декодирования/логирования payload values.
- PEM private key blocks.
- Version-pinned common GitHub/GitLab/OpenAI/AWS-like token patterns.
- Credential-bearing DSN/connection strings с conservative parsing.
- Строковый `Action`: allow, mask, block; простой config per entity/family.
- Безопасная block result/error semantics без sensitive fragment.
- Resolver priorities и overlap с URL/EMAIL.
- Positive/negative/malformed corpus, redaction и concurrency tests.

## Out of scope

- Entropy detector, semantic classifier, live token verification и provider APIs.
- YAML/rego/policy DSL, rules service и tenant-aware policy.
- Secret rotation, revocation и persistent incident store.

## Планируемые задачи OpenSpec

- [ ] Зафиксировать supported token versions и action semantics.
- [ ] Реализовать JWT и PEM detectors.
- [ ] Реализовать provider token patterns с false-positive boundaries.
- [ ] Реализовать credential-bearing DSN detector.
- [ ] Реализовать minimal action configuration и block path.
- [ ] Обновить resolver/Mask behavior для secret actions.
- [ ] Добавить corpus, safe-error, overlap и concurrency tests.
- [ ] Документировать pattern maintenance и limitations.

## Acceptance

1. Каждый supported secret family имеет versioned positive/negative cases и
   проходит общий validation/resolver pipeline.
2. Default/configured secret action однозначен; block не возвращает masked text как
   будто запрос безопасно продолжать.
3. Errors, audit-ready events и formatting не содержат raw secret или JWT payload.
4. DSN маскируется только при credential components; обычный URL не становится
   connection secret.
5. Core не получает external/network/policy-framework dependency.

## Verification

```bash
go test ./... -run 'Test(JWT|PEM|APIKey|Secret|DSN|Policy|Block)'
go test ./...
go vet ./...
go test -race ./...
openspec validate --specs --strict --no-interactive
```

Каждый secrets fuzz target запускается отдельно по точному package/name,
зафиксированному в OpenSpec tasks и task packet.

## Exit evidence

- Archived `secret-detection` и `minimal-policy` specs.
- Version/source note для provider patterns.
- Review evidence raw-secret leakage surfaces.
